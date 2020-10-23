# lib-vault

This lib provides some methods to deal with vault's golang client

## How it works

Implement new client
```
client, err := CreateClient()
```

Read secret
```
client.ReadSecret("path", "field")
```

List secret paths
```
client.ListSecret("basepath")
```

