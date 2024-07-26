package libvault

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"

	vault "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

func CreateClient() (*vault.Client, error) {
	// use vault token from env in priority like official cli
	vaultToken, tokenExists := os.LookupEnv("VAULT_TOKEN")
	if !tokenExists {
		homePath, err := homedir.Dir()
		if err != nil {
			return nil, errors.Wrap(err, "error getting user's home directory")
		}
		vaultTokenFile := filepath.Join(homePath, ".vault-token")
		if _, err = os.Stat(vaultTokenFile); err == nil {
			content, err := os.ReadFile(vaultTokenFile)
			if err != nil {
				return nil, errors.Wrap(err, "unable to read .vault-token file")
			}

			vaultToken = string(content)
			if vaultToken == "" {
				return nil, errors.New("no token found in your .vault-token file")
			}
		}
	}

	if vaultToken == "" {
		return nil, errors.New("Couldn't find neither $VAULT_TOKEN nor ~/.vault-token file")
	}

	client, err := vault.NewClient(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init vault client")
	}
	client.SetToken(vaultToken)

	return client, nil
}

func CreateClientWithAppRole(roleID, secretID string) (*vault.Client, error) {
	client, err := vault.NewClient(nil)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to initialize Vault client")
	}

	data := map[string]interface{}{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	resp, err := client.Logical().Write("auth/approle/login", data)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to generate token")
	}

	if resp.Auth == nil {
		return nil, errors.New("no authentication info returned")
	}

	client.SetToken(resp.Auth.ClientToken)
	return client, nil
}

func GetTokenTtlLeft(client *vault.Client) (int, error) {
	secret, err := client.Auth().Token().LookupSelf()
	if err != nil {
		return 0, errors.Wrap(err, "failed to lookup token")
	}

	return int(secret.Data["ttl"].(float64)), nil
}
