package vaultv2

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/mirakl/lib-vault/v2/internal/libvault"
	"github.com/mitchellh/go-homedir"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

var (
	v2Endpoint string
)

// Launch a docker with Vault in KV v2
func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	options := &dockertest.RunOptions{
		Repository: "vault",
		Tag:        "latest",
		Env:        []string{"VAULT_DEV_ROOT_TOKEN_ID=root"},
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(options)
	if err != nil {
		//resource.GetPort(resource)
		log.Fatalf("Could not start resource: %s", err)
	}

	v2Endpoint = fmt.Sprintf("http://localhost:%s", resource.GetPort("8200/tcp"))
	code := m.Run()
	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func addSecret(t *testing.T, client *Client, path string, values map[string]interface{}) {
	_, err := client.Client.Logical().Write(path, values)
	require.NoError(t, err)
}

func TestCreateClientTokenFromEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("VAULT_ADDR", v2Endpoint)
	os.Setenv("VAULT_TOKEN", "root")

	_, err := CreateClient()
	require.NoError(t, err, "This should not fail")
}

func TestCreateClientTokenFromFile(t *testing.T) {
	os.Clearenv()
	os.Setenv("VAULT_ADDR", v2Endpoint)

	_, err := CreateClient()
	require.Error(t, err, "Couldn't find neither $VAULT_TOKEN nor ~/.vault-token file")

	tmpdir, err := ioutil.TempDir("", "vaultread_test")
	if err != nil {
		t.Fatalf("unable to create tmpdir %q : %q", tmpdir, err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tmpdir)
	})

	os.Setenv("HOME", tmpdir)
	homePath, err := homedir.Dir()
	if err != nil {
		t.Fatalf("error getting user's home directory %q : %q", homePath, err)
	}
	tokenPath := filepath.Join(homePath, ".vault-token")
	if err := ioutil.WriteFile(tokenPath, []byte("root"), 0600); err != nil {
		t.Fatalf("unable to write file : %q", tokenPath)
	}

	_, err = libvault.CreateClient()
	require.NoError(t, err, "This should not fail")
}

func TestReadSecretV2(t *testing.T) {
	os.Clearenv()
	os.Setenv("VAULT_ADDR", v2Endpoint)
	os.Setenv("VAULT_TOKEN", "root")

	client, err := CreateClient()
	require.NoError(t, err, "This should not fail")

	addSecret(t, client, "secret/data/foo", map[string]interface{}{
		"data": map[string]interface{}{
			"secret": "bar",
		},
	})

	s, err := client.ReadSecret("secret/foo", "secret")
	require.NoError(t, err)
	require.Equal(t, "bar", s)

	_, err = client.ReadSecret("secret/anything", "secret")
	require.Error(t, err, "no exist secret \"secret/anything\"")
}

func TestListSecretV2(t *testing.T) {
	os.Clearenv()
	os.Setenv("VAULT_ADDR", v2Endpoint)
	os.Setenv("VAULT_TOKEN", "root")

	client, err := CreateClient()
	require.NoError(t, err, "This should not fail")

	addSecret(t, client, "secret/data/foo", map[string]interface{}{
		"data": map[string]interface{}{
			"secret": "bar",
		},
	})

	addSecret(t, client, "secret/data/foo1", map[string]interface{}{
		"data": map[string]interface{}{
			"secret": "bar",
		},
	})

	addSecret(t, client, "secret/data/foo2", map[string]interface{}{
		"data": map[string]interface{}{
			"secret": "bar",
		},
	})

	secrets, err := client.ListSecretPath("secret")
	require.NoError(t, err)
	require.Equal(t, 3, len(secrets))

	require.Equal(t, []string{"secret/foo", "secret/foo1", "secret/foo2"}, secrets)

	_, err = client.ListSecretPath("secret/not_exist")
	require.Error(t, err, "The path \"secret/not_exist\" does not exist")
}

func TestGetSecretKvV2(t *testing.T) {
	os.Clearenv()
	os.Setenv("VAULT_ADDR", v2Endpoint)
	os.Setenv("VAULT_TOKEN", "root")

	client, err := CreateClient()
	require.NoError(t, err, "This should not fail")

	addSecret(t, client, "secret/data/foo", map[string]interface{}{
		"data": map[string]interface{}{
			"secret":  "bar",
			"secret2": "bar2",
			"secret3": "bar3",
		},
	})

	s, err := client.GetSecret("secret/foo")
	require.NoError(t, err)

	require.Equal(t, "bar", s["secret"])
	require.Equal(t, "bar2", s["secret2"])
	require.Equal(t, "bar3", s["secret3"])
}
