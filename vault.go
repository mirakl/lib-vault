package lib_vault

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

type VaultClient struct {
	client *vault.Logical
}

func CreateClient() (*VaultClient, error) {
	// use vault token from env in priority like official cli
	vaultToken, tokenExists := os.LookupEnv("VAULT_TOKEN")
	if !tokenExists {
		homePath, err := homedir.Dir()
		if err != nil {
			return nil, errors.Wrap(err, "error getting user's home directory")
		}
		vaultTokenFile := filepath.Join(homePath, ".vault-token")
		if _, err = os.Stat(vaultTokenFile); err == nil {
			content, err := ioutil.ReadFile(vaultTokenFile)
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

	return &VaultClient{client.Logical()}, nil
}

func (vc *VaultClient) ListSecret(path string) ([]string, error) {
	s, err := vc.client.List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list vault secrets: %w", err)
	}

	secrets := s.Data["keys"]

	var result []string
	t, ok := secrets.([]interface{})
	if !ok {
		return nil, errors.New("Incompatible type")
	}

	for _, value := range t {
		f := fmt.Sprintf("%v/%v", path, value)
		result = append(result, f)
	}

	return result, nil
}

func (vc *VaultClient) ReadSecret(path string, field string) (string, error) {
	secret, err := vc.client.Read(path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read secret %q", path)
	}
	if secret == nil {
		return "", errors.Errorf("no exist secret %q", path)
	}

	value, found := secret.Data[field]
	if !found {
		return "", errors.Errorf("no field %q for secret %q", field, secret.WrapInfo.CreationPath)
	}

	convertedValue, ok := value.(string)
	if !ok {
		return "", errors.Errorf("secret %q in %q has type %T = %v", field, secret.WrapInfo.CreationPath, value, value)
	}
	if convertedValue == "" {
		return "", errors.Errorf("value is empty for field %q in %q", field, secret.WrapInfo.CreationPath)
	}

	return convertedValue, nil
}
