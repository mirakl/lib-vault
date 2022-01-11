package libvault

import (
	"io/ioutil"
	"os"
	"path/filepath"

	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"
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

	return client, nil
}
