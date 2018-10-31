# Mutating Webhook Vault Agent

## Build Vault Agent Webhook

```
    cd openshift/

    oc project hashicorp-vault

    oc create -f build/go-dep-build.yaml
    oc create -f build/vault-agent-webhook-build.yaml
```

## Deploy Vault Agent WebHook

1. Configuration

    ```
    oc create -f template/webhook-configmap.yaml
    oc create -f template/vault-agent-configmap.yaml
    oc create -f template/webhook-service.yaml
    oc create -f template/webhook-podsecuritypolicy.yaml
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

1. Label the project where you want the webhook to listen.

    ```
    oc label namespace app vault-agent-webhook=enabled
    ```

2. Add the *sidecar.agent.vaultproject.io/inject* annotation with value true to the pod template spec to enable injection.


# References

* https://docs.openshift.com/container-platform/3.9/architecture/additional_concepts/dynamic_admission_controllers.html
* https://v1-9.docs.kubernetes.io/docs/admin/extensible-admission-controllers/
* https://github.com/morvencao/kube-mutating-webhook-tutorial/
* https://github.com/kubernetes/kubernetes/tree/release-1.9/test/images/webhook
