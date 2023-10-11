Kubernetes Mutating Webhook
====

This project is a [Kubernetes Mutating Admission Webhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/)
allowing manipulation of a Pod *image*, _pullSecrets_ and _pullPolicy_, 
and PersistentVolumeClaim storageClassName

1) Pod image can be prepended with a given registry
2) An `imagePullSecrets` can be injected in Pod spec
3) `imagePullPolicy` can be forced to _Always_
4) An `storageClassName` can be forced to PersistentVolumeClaim objects.

# Rationale 

## Force image registry and injecting `imagePullSecrets`

In Sqooba's Kubernetes deployment blueprint, we want to control image provenance
via a central container registry (acting as either cache or proxy). To really enforce
all the images come from this given container registry, this mutating webhook
is prepend'ing (concatenating at the begining of the image name) 
a given registry to the image name. See below for more details
about the heuristic used to achieve this.

## Force image pull policy to Always

Attentive readers might wonder why this hand made admission webhook has been written
while this feature already exists as a builtin admission controller with name `AlwaysPullImages`?

At the time of writing, [EKS 1.13.7](https://docs.aws.amazon.com/eks/latest/userguide/platform-versions.html)
doesn't have this builtin admission controller enabled,
hence if you want this behavior in your EKS cluster, you need to rewrite it
and to deploy as a regular mutating webhook.

The security need of having the `AlwaysPullImages` admission plugin enabled is detailed here:
https://medium.com/@trstringer/kubernetes-alwayspullimages-admission-control-the-importance-implementation-and-security-d83ff3815840

Also, if you're running kubernetes locally for the development, you might want to use
locally build images, which, if you're using the tag `latest`, requires to force the
image pull policy to `Never` (as a reminder, `latest` has a special meaning of setting the
default pull policy to `Always` instead of `IfNotPresent`).
 
# Configuration

In the `deployment.yaml` file generated via the provided command (details below),
few environment variables drive the configuration:

| Environment variable         | Default  | Description                                                                                                                                                                                                     |
|------------------------------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `REGISTRY`                   |          | If set, tells which registry to force, such as `docker.sqooba.io`                                                                                                                                               |
| `IMAGE_PULL_SECRET`          |          | If set, tells which `imagePullSecrets` to inject in the Pod. Note the secret must be present in the namespace, and this task is out of this webhook responsibility.                                             |
| `IMAGE_PULL_SECRET_APPEND`   | `false`  | Tells whether the `IMAGE_PULL_SECRET` is appen'ed to an existing list of `imagePullSecrets` (if it does not exist already) or if any `imagePullSecrets` are replaced by `IMAGE_PULL_SECRET` (default behavior). |
| `FORCE_IMAGE_PULL_POLICY`    |          | If set to true, `imagePullPolicy` will be forced to the value of `IMAGE_PULL_POLICY_TO_FORCE`                                                                                                                   |
| `IMAGE_PULL_POLICY_TO_FORCE` | `Always` | The `imagePullPolicy` to set.                                                                                                                                                                                   |
| `DEFAULT_STORAGE_CLASS`      |          | If set, enforce storage class of PVCs to the value, such as `rook-ceph-block`, if no other storage class is set.                                                                                                |
| `EXCLUDE_NAMESPACES`         |          | Optional list, comma separated, of namespace(s) to exclude, for instance "kube-system,default". To keep the behavior backward compatible, set this value to `kube-system,kube-public`                           |
| `IGNORED_REGISTRIES`         |          | Optional list, comma separated, of registries that should be ignored by the webhook (besides the one specified via the REGISTRY parameter)                                                                      |
| `LOG_LEVEL`                  | `info`   | This option lets you define a logging verbosity between trace, debug, info (the default), warn, error or fatal.                                                                                                 |

# Image registry heuristic

A heuristic is used to determine if the current image already
contains a registry as part of its name.

An container image is expected to be of the following format:

```
[registry[:port]/]?[classifier/]*image[:tag[@sha256:...]]?
```

- An optional registry, with an optional port
- Zero or more classifier
- A mandatory image name
- An optional tag, separated from the image name via `:`, which can contains hash

The difference between a *registry* and a *classifier* is that a registry contains
one or more `.`, for instance `a.b.c`, where classifier doesn't.

Rewriting rule can be expressed as follow:
1) If a registry is present, it is replaced by the one given as parameter.
2) If no registry are found using the heuristic above, it is prepended.

Example: Let's assume the registry is `r`

- `a:v` -> `r/a:v`
- `a/b:v` -> `r/a/b:v`
- `a.b/c:v` -> `r/c:v`
- `a.b:v` -> invalid image definition, but still replaced to `r:e`

# Acknowledgements

This project takes high inspiration from [https://github.com/stackrox/admission-controller-webhook-demo](https://github.com/stackrox/admission-controller-webhook-demo)

Some more inspiration has been taken directly from Kubernetes `alwayspullimages` 
[admission plugin](https://github.com/kubernetes/kubernetes/blob/master/plugin/pkg/admission/alwayspullimages/admission.go)

# Install the webhook

A script, [deployment/generate-certs-and-config.sh](deployment/generate-certs-and-config.sh)
is provided and, as the names says, it generates TLS certificates for the webhook and
is templating few manifest files to include all the required details. Resulting, ready-to-be-used
manifests files will be output'ed in `deployment/generated-yyyymmdd` folder.

```
cd deployment
./generate-certs-and-config.sh
```

Then deploy the manifest in your favourite K8s cluster:

```
kubectl apply -f generated-yyyymmdd/
```

# Test

Let's deploy a dummy pod with a image pointing to the central docker hub
see what is happening:

```
apiVersion: v1
kind: Pod
metadata:
  name: mutatingwebhookpodtest
spec:
  containers:
  - name: busybox
    image: busybox:1.28
    command:
      - sleep
      - "3600"
```

After the pod has been deployed using `kubectl` command, one can ensure that the `image` has
been appended to the registry

```
kubectl get po ${pod-name} -o jsonpath="{.spec.containers[0].image}" 
```

# Hack

```
make all
```

# More pointers

- https://docs.giantswarm.io/guides/creating-your-own-admission-controller/
- https://github.com/open-policy-agent/gatekeeper
- https://github.com/giantswarm/grumpy/blob/instance_migration/grumpy.go

# FAQ

## Upgrade

Updating a minor version of the webhook

```
kubectl -n kube-system set \
    image deploy/k8s-mutate-image-and-policy-webhook \
    webhook=sqooba/k8s-mutate-image-and-policy-webhook:3.2.2
```

# Restart

Restarting the webhook

```
kubectl -n kube-system rollout restart deploy/k8s-mutate-image-and-policy-webhook
```

# Disable

In case of any problem, delete the mutation webhook, fix, and re-add.

```
kubectl delete MutatingWebhookConfiguration k8s-mutate-image-and-policy-webhook
```
