FROM gcr.io/distroless/base

COPY k8s-mutate-image-and-policy-webhook /k8s-mutate-image-and-policy-webhook

USER nobody

ENTRYPOINT ["/k8s-mutate-image-and-policy-webhook"]

HEALTHCHECK --interval=30s --timeout=3s CMD ["/k8s-mutate-image-and-policy-webhook", "--health-check"]

EXPOSE 8443
