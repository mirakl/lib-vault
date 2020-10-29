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
client.ListSecrets("secret/kube/clients")

=>
 secret/kube/clients/gke-prod-ae-1-kubeconfig
 secret/kube/clients/gke-prod-sae-1-credentials
 secret/kube/clients/gke-test-eu-1-kubeconfig
 secret/kube/clients/kube-preprod-eu-1-kubeconfig
 secret/kube/clients/kube-test-eu-1-kubeconfig
 ...

```

