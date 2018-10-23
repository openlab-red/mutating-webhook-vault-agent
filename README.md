# Mutating Webhook Vault Agent

## Build Vault Agent Webhook

```
    oc new-build --name=vault-agent-webhook https://github.com/openlab-red/mutating-webhook-vault-agent
```

## Deploy Vault Agent WebHook

1. Configuration

    ```
    oc create -f template/webhook-configmap.yaml
    oc create -f template/vault-agent-configmap.yaml
    oc create -f template/webhook-service.yaml
    ```

2. Deployment

    ```
    oc create -f template/webhook-deployment.yaml
    ```

3. Create Mutating WebHook


    ```
    oc create -f template/webhook-mutating-config.yaml
    ```

## Verify Injection

WIP

```
oc label namespace app vault-agent-webhook=enabled
```