package v2

import (
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/mirakl/lib-vault/libvault"
	"github.com/pkg/errors"
)

type v2Client struct {
	Client *vault.Client
}

func CreateClient() (*v2Client, error) {
	client, err := libvault.CreateClient()
	if err != nil {
		return nil, errors.Wrapf(err, "")
	}

	return &v2Client{
		Client: client,
	}, nil
}

func (vc *v2Client) ReadSecret(path string, field string) (string, error) {
	secret, err := vc.GetSecret(path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read secret in %q", path)
	}

	convertedValue, ok := secret[field].(string)
	if !ok {
		return "", errors.Errorf("field %q does not exist for this secret %q", field, path)
	}

	return convertedValue, nil
}

func (vc *v2Client) GetSecret(path string) (map[string]interface{}, error) {
	v2Path := kvV2Path(path, "data")
	secret, err := vc.Client.Logical().Read(v2Path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read secret in %q", path)
	}

	if secret == nil {
		return nil, errors.Errorf("No secret exists in %q", path)
	}

	m, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("Incompatible type %q key does not exist, %T %#v", "data", secret.Data["data"], secret.Data["data"])
	}

	return m, nil
}

func (vc *v2Client) ListSecretPath(path string) ([]string, error) {
	v2Path := kvV2Path(path, "metadata")

	s, err := vc.Client.Logical().List(v2Path)
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

func kvV2Path(path string, key string) string {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	return strings.Replace(path, "/", "/"+key+"/", 1)
}
