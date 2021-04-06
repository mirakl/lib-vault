# lib-vault

This lib provides some methods to deal with vault's (kv v1 or v2) golang client

## How it works

Create new client, the creation depends on existing of either  **VAULT_TOKEN** environment variable 
or `.vault-token` file in your home directory 
```
client, err := CreateClient()
```

Read secret (kv v1), returns the secret value
```
client.ReadSecret("path", "field")
```

List secret paths (kv v1), returns absolute secret path from base path.
```
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

Read secret (kv v2), returns the secret value
```
client.ReadSecretKvV2("path", "field")
```

List secret paths (kv v2), returns absolute secret path from base path.
```
client.ListSecretPathKvV2("basepath")
```

example:
```
client.ListSecretPathKvV2("secret/foo/bar")

 secret/foo/bar/foo1
 secret/foo/bar/foo2
 secret/foo/bar/foo3
 secret/foo/bar/foo4
 secret/foo/bar/foo5
 ...

```
