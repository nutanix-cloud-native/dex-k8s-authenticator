FROM gcr.io/distroless/static@sha256:4c2cd1dee442c9a99924a378c7c126fac8fef14771579ef567c3f2bdd7a5122a

COPY ./bin/linux/amd64/dex-k8s-authenticator /app/bin/dex-k8s-authenticator
COPY html /app/html
COPY templates /app/templates
COPY templates /app/original-templates

WORKDIR /app

ENTRYPOINT ["/app/bin/dex-k8s-authenticator"]
CMD ["--help"]
