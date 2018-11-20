# Mutating Webhook Vault Agent

## Build Vault Agent Webhook container

```
    oc project hashicorp-vault

    oc apply -f openshift/webhook-build.yaml
```

## Deploy Vault Agent WebHook

1. Create the sidecar vault agent configuration

    ```
    oc apply -f openshift/sidecar-configmap.yaml
    ```

2. Process Mutating WebHook Template.
   
   The template is going to create the following resources:
    * vault-agent-webhook-psp PodSecurityPolicy
    * vault-agent-webhook-clusterrole ClusterRole
    * vault-agent-webhook ServiceAccount
    * vault-agent-webhook-rolebinding ClusterRoleBinding
    * vault-agent-webhook Service
    * vault-agent-webhook DeploymentConfig
    * vault-agent-webhook MutatingWebhookConfiguration
    
   2.1 Get service-ca.crt from the vault pod

    ```
    pod=$(oc get pods -lapp=vault --no-headers -o custom-columns=NAME:.metadata.name)
    export CA_BUNDLE=$(oc exec $pod -- cat /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt | base64 | tr -d '\n')
    ```

   2.2 Process the webhook-template

    ```
    oc process -f openshift/webhook-template.yaml -p CA_BUNDLE=${CA_BUNDLE} | oc apply -f -
    ```

    |     PARAMETER   |  DEFAULT           |  DESCRIPTION                                                              |
    |-----------------|--------------------|---------------------------------------------------------------------------|
    | CA_BUNDLE       |                    |    CA used by kubernetes to trust the webhook                             |
    | VAULT_NAMESPACE |    hashicorp-vault |    Hashicorp Vault Namespac                                               |
    | GIN_MODE        |    release         |    Http server startup mode [gin-gonic](https://github.com/gin-gonic/gin) |
    | LOG_LEVEL       |    INFO            |    Log level from [logrus](https://github.com/sirupsen/logrus)            |

## Verify Sidecar Injection

1. Label the target project where you want the webhook to inject the vault agent sidecar container.

    ```
    oc label namespace app vault-agent-webhook=enabled
    ```

2. Add the *sidecar.agent.vaultproject.io/inject* annotation with value true to the pod template spec to enable injection.


    ```
    oc patch dc/thorntail-example -p '{
                                     "spec": {
                                       "template": {
                                         "metadata": {
                                           "annotations": {
                                             "sidecar.agent.vaultproject.io/inject": "true",
                                             "sidecar.agent.vaultproject.io/secret-key": "secret/example",
                                             "sidecar.agent.vaultproject.io/properties-ext": "yaml",
                                             "sidecar.agent.vaultproject.io/vault-role": "example"
                                           }
                                         }
                                       }
                                     }
                                   }'
    ```
3. The vault agent webhook will:
    * Create or Update the vault agent configmap
    * Inject Vault agent sidecar container
    * Inject Vault secret fetcher sidecar container
    * Mount Vault volume to the app container

# References

* https://docs.openshift.com/container-platform/3.10/architecture/additional_concepts/dynamic_admission_controllers.html
* https://v1-10.docs.kubernetes.io/docs/admin/extensible-admission-controllers/
* https://github.com/morvencao/kube-mutating-webhook-tutorial/
* https://github.com/kubernetes/kubernetes/tree/release-1.10/test/images/webhook
