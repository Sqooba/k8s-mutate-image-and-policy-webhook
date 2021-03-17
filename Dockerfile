FROM gcr.io/distroless/base

COPY k8s-mutate-image-and-policy-webhook /k8s-mutate-image-and-policy-webhook

USER nobody

ENTRYPOINT ["/k8s-mutate-image-and-policy-webhook"]

EXPOSE 8443
