package vaultv1

import (
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"github.com/mirakl/lib-vault/internal/libvault"
	"github.com/pkg/errors"
)

type Client struct {
	Client *vault.Client
}

func CreateClient() (*Client, error) {
	client, err := libvault.CreateClient()
	if err != nil {
		return nil, errors.Wrapf(err, "")
	}

	return &Client{
		Client: client,
	}, nil
}

func (vc *Client) ListSecretPath(path string) ([]string, error) {
	s, err := vc.Client.Logical().List(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list vault secrets")
	}

	if s == nil {
		return nil, errors.Errorf("The path %q does not exist", path)
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

func (vc *Client) ReadSecret(path string, field string) (string, error) {
	secret, err := vc.GetSecret(path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read secret in %q", path)
	}

	value, found := secret[field]
	if !found {
		return "", errors.Errorf("no field %q for secret in %q", field, path)
	}

	convertedValue, ok := value.(string)
	if !ok {
		return "", errors.Errorf("secret %q in %q has type %T = %v", field, path, value, value)
	}
	if convertedValue == "" {
		return "", errors.Errorf("value is empty for field %q in %q", field, path)
	}

	return convertedValue, nil
}

func (vc *Client) GetSecret(path string) (map[string]interface{}, error) {
	secret, err := vc.Client.Logical().Read(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read secret in %q", path)
	}

	if secret == nil {
		return nil, errors.Errorf("no existing secret in %q", path)
	}

	return secret.Data, nil
}
