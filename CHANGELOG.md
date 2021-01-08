k8s-mutate-image-and-policy
====

# Version v3.1.0 -- 08.12.2021

## Enhancement

- Move to go mode instead of go dep
- Change from DEBUG and TRACE variable to proper LOG_LEVEL
- Add `/debug/verbosity` endpoint to allow change of log level at runtime
- Run a nobody user instead of root

## Bug fix

- support port in registry, i.e. my.private.registry:5000.

# Version 3.0.0 -- 08.08.2020

## Breaking change

- Change to admissionregistration.k8s.io/v1. This requires kubernetes v1.16 or above, and a complete redeployment of the webhook.
- k8s-mutate-image-and-policy deployment name has been changed in k8s-mutate-image-and-policy-webhook

## Bug fix

- support Skaffold image names, i.e. with sha256.

# Version 2.3.1 -- 29.11.2019

## Bug fix

- Handle exotic registries with ports, such as x.y:80
- Handle images only with dots (in the version)

# Version 2.3.0 -- 19.11.2019

## Improvement

- Supports registry with a specific port, such a.b:80. 
