k8s-mutate-image-and-policy
====

# Version v3.4.0 -- unreleased

## Enhancement

- Allow to append ImagePullSecret to an existing list via new flag `IMAGE_PULL_SECRET_APPEND`
- Upgrade to Goland 1.20
- Allow to specify which pull policy (Always, IfNotPresent or Never) to force, via new flag `IMAGE_PULL_POLICY_TO_FORCE`

## Change

- Avoid mutating `ImagePullSecrets` when the expected value is already present.
- golangci-lint'ing, goimports'ing and gofmt'ing

## Other

- Golang v1.19, k8s version bump

# Version v3.3.1 -- 23.11.2021

## Other

- Build image via Github workflow

# Version v3.3.0 -- 16.09.2021

## Enhancement

- Add `IGNORE_REGISTRIES` option
- Build multi-arch image

# Version v3.2.2 -- 27.05.2021

## Fix

- Insert git hash, build date and version at build time

## Other

- Code simplification (healthcheck has been removed, since it's not working with https...)

# Version v3.2.1 -- 14.03.2021

## Enhancement

- Bump/clean/purge dependencies

# Version v3.2.0 -- 12.03.2021

## Enhancement

- Set subjecAltname in the certificat (k8s 1.20, go 1.15)
- Set default TLS_CERT_FILE and TLS_KET_FILE
- Bump sqooba/go-common version

# Version v3.1.0 -- 08.12.2020

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
