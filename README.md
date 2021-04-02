# lib-vault




This lib provides some methods to deal with vault's golang client

## How it works

Create new client, the creation depends on existing of either  **VAULT_TOKEN** environment variable 
or `.vault-token` file in your home directory 
```
client, err := CreateClient()
```

Read secret, returns the secret value
```
client.ReadSecret("path", "field")
```

List secret paths, returns absolute secret path from base path.
```
client.ListSecrets("basepath")
```

example:
```
client.ListSecrets("secret/foo/bar")

 secret/foo/bar/foo1
 secret/foo/bar/foo2
 secret/foo/bar/foo3
 secret/foo/bar/foo4
 secret/foo/bar/foo5
 ...

```

