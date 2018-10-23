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

    3.1 Get service-ca.crt

        ```
        pod=$(oc get pods -lapp=vault-agent-webhook --no-headers -o custom-columns=NAME:.metadata.name)
        export CA_BUNDLE=$(oc exec $pod -- cat /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt | base64 | tr -d '\n')
        ```

    3.2 Create the webhook

        ```
        oc process -f template/webhook-mutating-config.yaml -p CA_BUNDLE=${CA_BUNDLE} | oc create -f -
        ```

## Verify Injection

WIP

```
oc label namespace app vault-agent-webhook=enabled
```