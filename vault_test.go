package lib_vault

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	"github.com/mitchellh/go-homedir"
)

func createTestVault(t *testing.T) (string, string) {
	t.Helper()

	// Create an in-memory, unsealed core (the "backend", if you will).
	core, keyShares, rootToken := vault.TestCoreUnsealed(t)
	_ = keyShares

	// Start an HTTP server for the core.
	ln, addr := http.TestServer(t, core)
	t.Cleanup(func() {
		if err := ln.Close(); err != nil {
			t.Log(err)
		}
	})

	// Create a client that talks to the server, initially authenticating with
	// the root token.
	conf := api.DefaultConfig()
	conf.Address = addr

	client, err := api.NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}
	client.SetToken(rootToken)

	// Setup required secrets, policies, etc.
	_, err = client.Logical().Write("secret/foo", map[string]interface{}{
		"secret": "bar",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Logical().Write("secret/foo1", map[string]interface{}{
		"secret": "bar",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Logical().Write("secret/foo2", map[string]interface{}{
		"secret": "bar",
	})
	if err != nil {
		t.Fatal(err)
	}

	return addr, rootToken
}

func TestCreateClient(t *testing.T) {
	addr, token := createTestVault(t)

	os.Clearenv()
	os.Setenv("VAULT_ADDR", addr)
	os.Setenv("VAULT_TOKEN", token)

	_, err := CreateClient()
	require.NoError(t, err, "This should not fail")

	os.Clearenv()
	os.Setenv("VAULT_ADDR", addr)

	tmpdir, err := ioutil.TempDir("", "vaultread_test")
	if err != nil {
		t.Fatalf("unable to create tmpdir %q : %q", tmpdir, err)
	}

	t.Cleanup(func() {
		removeTmpDir(t, tmpdir)
	})

	os.Setenv("HOME", tmpdir)
	homePath, err := homedir.Dir()
	if err != nil {
		t.Fatalf("error getting user's home directory %q : %q", homePath, err)
	}
	tokenPath := filepath.Join(homePath, ".vault-token")
	if err := ioutil.WriteFile(tokenPath, []byte(token), 0600); err != nil {
		t.Fatalf("unable to write file : %q", tokenPath)
	}

	_, err = CreateClient()
	require.NoError(t, err, "This should not fail")

	removeTmpDir(t, tmpdir)

	_, err = CreateClient()
	require.Error(t, err, "Couldn't find neither $VAULT_TOKEN nor ~/.vault-token file")

}

func removeTmpDir(t *testing.T, tmpdir string) {
	if err := os.RemoveAll(tmpdir); err != nil {
		t.Log(err)
	}
}

func TestReadSecret(t *testing.T) {
	addr, token := createTestVault(t)

	os.Clearenv()
	os.Setenv("VAULT_ADDR", addr)
	os.Setenv("VAULT_TOKEN", token)

	client, err := CreateClient()
	require.NoError(t, err, "This should not fail")

	s, err := client.ReadSecret("secret/foo", "secret")
	require.NoError(t, err)
	require.Equal(t, "bar", s)

	_, err = client.ReadSecret("secret/anything", "secret")
	require.Error(t, err, "no exist secret \"secret/anything\"")
}

func TestListSecret(t *testing.T) {
	addr, token := createTestVault(t)

	os.Clearenv()
	os.Setenv("VAULT_ADDR", addr)
	os.Setenv("VAULT_TOKEN", token)

	client, err := CreateClient()
	require.NoError(t, err, "This should not fail")

	secrets, err := client.ListSecrets("secret")
	require.NoError(t, err)
	require.Equal(t, 3, len(secrets))

	require.Equal(t, []string{"secret/foo", "secret/foo1", "secret/foo2"}, secrets)
}
