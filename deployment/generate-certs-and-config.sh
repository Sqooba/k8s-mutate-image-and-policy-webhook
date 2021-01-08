#!/usr/bin/env bash

# generate-keys.sh
#
# Generate a (self-signed) CA certificate and a certificate and private key to be used by the webhook server.
# The certificate will be issued for the Common Name (CN) of `k8s-mutate-image-and-policy.kube-system.svc`, which is the
# cluster-internal DNS name for the service.

export NAMESPACE=kube-system

mkdir -p generated
mkdir certs

chmod 0700 "certs"
pushd "certs"

# Generate the CA cert and private key
openssl req -days 730 -nodes -new -x509 -keyout ca.key -out ca.crt -subj "/CN=Sqooba k8s-mutate-image-and-policy-webhook"
# Generate the private key for the webhook server
openssl genrsa -out webhook-server-tls.key 2048
# Generate a Certificate Signing Request (CSR) for the private key, and sign it with the private key of the CA.
openssl req -new -key webhook-server-tls.key -subj "/CN=k8s-mutate-image-and-policy-webhook.${NAMESPACE}.svc" \
    | openssl x509 -req -days 730 -CA ca.crt -CAkey ca.key -CAcreateserial -out webhook-server-tls.crt

kubectl create secret generic k8s-mutate-image-and-policy-webhook-tls-certs -n ${NAMESPACE} \
    --from-file=./ca.crt \
    --from-file=./webhook-server-tls.key \
    --from-file=./webhook-server-tls.crt \
    --dry-run -o yaml > ../generated/certs-configmap.yaml

popd

export CA_PEM_B64=$(cat certs/ca.crt | base64)

cat deployment.yaml.tmpl | envsubst > generated/deployment.yaml

mv generated "generated-$(date +%Y%m%d)"

echo "Deployment configuration has been successfully generated in generated-$(date +%Y%m%d)"
echo "You can now safely call kubectl apply -f generated-$(date +%Y%m%d)/"
