package vaultv1_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	vault "github.com/hashicorp/vault/api"
	"github.com/mirakl/lib-vault/v2/vaultv1"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/require"
)

var (
	v1Endpoint string
)

// Launch a docker with Vault in KV v1
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

	v1Endpoint = fmt.Sprintf("http://localhost:%s", resource.GetPort("8200/tcp"))
	fmt.Println("TestMAin baby")
	code := m.Run()
	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func addSecret(t *testing.T, client *vaultv1.Client, path string, values map[string]interface{}) {
	_, err := client.Client.Logical().Write(path, values)
	require.NoError(t, err)
}

func TestReadSecret(t *testing.T) {
	os.Clearenv()
	os.Setenv("VAULT_ADDR", v1Endpoint)
	os.Setenv("VAULT_TOKEN", "root")

	client, err := vaultv1.CreateClient()
	require.NoError(t, err, "This should not fail")
	err = client.Client.Sys().Mount("kv-v1", &vault.MountInput{
		Type:    "kv",
		Options: map[string]string{"version": "1"},
	})
	require.NoError(t, err, "This should not fail")
	addSecret(t, client, "kv-v1/foo", map[string]interface{}{"secret": "bar"})

	s, err := client.ReadSecret("kv-v1/foo", "secret")
	require.NoError(t, err)
	require.Equal(t, "bar", s)

	_, err = client.ReadSecret("kv-v1/anything", "secret")
	require.Error(t, err, "no exist secret \"kv-v1/anything\"")
}

func TestListSecret(t *testing.T) {
	os.Clearenv()
	os.Setenv("VAULT_ADDR", v1Endpoint)
	os.Setenv("VAULT_TOKEN", "root")

	client, err := vaultv1.CreateClient()
	require.NoError(t, err, "This should not fail")
	err = client.Client.Sys().Mount("kv-v1-2", &vault.MountInput{
		Type:    "kv",
		Options: map[string]string{"version": "1"},
	})
	require.NoError(t, err, "This should not fail")
	addSecret(t, client, "kv-v1-2/foo", map[string]interface{}{"secret": "bar"})
	addSecret(t, client, "kv-v1-2/foo1", map[string]interface{}{"secret": "bar"})
	addSecret(t, client, "kv-v1-2/foo2", map[string]interface{}{"secret": "bar"})

	secrets, err := client.ListSecretPath("kv-v1-2")
	require.NoError(t, err)
	require.Equal(t, 3, len(secrets))

	require.Equal(t, []string{"kv-v1-2/foo", "kv-v1-2/foo1", "kv-v1-2/foo2"}, secrets)

	_, err = client.ListSecretPath("kv-v1-2/not_exist")
	require.Error(t, err, "The path \"kv-v1-2/not_exist\" does not exist")
}

func TestWithAppRole(t *testing.T) {
	os.Clearenv()
	os.Setenv("VAULT_ADDR", v1Endpoint)
	os.Setenv("VAULT_TOKEN", "root")

	vc, err := vaultv1.CreateClient()
	require.NoError(t, err)
	err = vc.Client.Sys().Mount("kv-v1-3", &vault.MountInput{
		Type:    "kv",
		Options: map[string]string{"version": "1"},
	})
	require.NoError(t, err, "This should not fail")

	addSecret(t, vc, "kv-v1-3/approle", map[string]interface{}{"roleID": "xxxxxx"})

	roleID, secretID := setupAppRole(t, vc)
	appRoleClient, err := vaultv1.CreateClientWithAppRole(fmt.Sprint(roleID), fmt.Sprint(secretID))
	require.NoError(t, err)

	secret, err := appRoleClient.GetSecret("kv-v1-3/approle")
	require.NoError(t, err)

	require.Equal(t, "xxxxxx", secret["roleID"])
}

func setupAppRole(t *testing.T, vc *vaultv1.Client) (string, string) {
	// Enable approle
	err := vc.Client.Sys().EnableAuthWithOptions("approle/", &vault.EnableAuthOptions{
		Type: "approle",
	})
	require.NoError(t, err)

	// Create a policy to allow the approle to do whatever
	err = vc.Client.Sys().PutPolicy("unittest", `
path "*" {
    capabilities = ["create", "read", "list", "update", "delete"]
}
`)
	require.NoError(t, err)

	// Create role
	_, err = vc.Client.Logical().Write("auth/approle/role/roletest", map[string]interface{}{
		"period":   "5m",
		"policies": []string{"unittest"},
	})
	require.NoError(t, err)

	// Get role_id
	resp, err := vc.Client.Logical().Read("auth/approle/role/roletest/role-id")
	require.NoError(t, err)
	roleID := resp.Data["role_id"]

	// Get secret_id
	resp, err = vc.Client.Logical().Write("auth/approle/role/roletest/secret-id", map[string]interface{}{})
	require.NoError(t, err)
	secretID := resp.Data["secret_id"]

	return fmt.Sprint(roleID), fmt.Sprint(secretID)
}
