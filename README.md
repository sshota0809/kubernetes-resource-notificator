# kubernetes-notificator

## How to Use

Sample manifests is in `manifest` directory. This controller is based on argocd-notifications, so check out [argocd-notifications Templates](https://argocd-notifications.readthedocs.io/en/stable/templates/) and [argocd-notifications Subscriptions](https://argocd-notifications.readthedocs.io/en/stable/subscriptions/).

## Options

```
-c, --configmap string               ConfigMap name for controller (default "kubernetes-notificator-cm")
-g, --group string                   apiGroup monitored by controller
-n, --namespace string               If present, the namespace scope for this CLI request
-r, --resource string                apiResource monitored by controller (default "pods")
-s, --secret string                  Secret name for controller (default "kubernetes-notificator-secret")
-v, --version string                 apiVersion monitored by controller (default "v1")
```

## Examples

### to Monitor Pod

```
--group ""
--version v1
--resource pods
```

### to Monitor Deployment

```
--group apps
--version v1
--resource deployments
```
