# lib-vault

This lib provides some methods to deal with vault's (kv v1 or v2) golang client

## How it works

Create new client, the creation depends on existing of either  **VAULT_TOKEN** environment variable
or `.vault-token` file in your home directory.

You have to choose between using vault KV API v1 or v2:

For KV v1:

```go

import "github.com/mirakl/lib-vault/v2/vaultv2"

client, err := vaultv1.CreateClient()
```
For KV v2:

```go

import "github.com/mirakl/lib-vault/v2/vaultv2"

client, err := vaultv2.CreateClient()
```

Both clients expose the same set of methods:


Read secret returns the secret value
```go
client.ReadSecret("path", "field")
```

Get secret returns the secret (map[string]interface{})
```go
client.GetSecret("path")
```

List secret paths returns absolute secret path from base path.
```go
client.ListSecretPath("basepath")
```


example:
```
client.ListSecretPath("secret/foo/bar")

 secret/foo/bar/foo1
 secret/foo/bar/foo2
 secret/foo/bar/foo3
 secret/foo/bar/foo4
 secret/foo/bar/foo5
 ...

```
