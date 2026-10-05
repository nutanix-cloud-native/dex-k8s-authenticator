FROM gcr.io/distroless/static@sha256:58133991db06659feaabe0f4e97a35cebf15ef4ea08f8a4c6d2ee5f75e4aa6a0

COPY ./bin/linux/amd64/dex-k8s-authenticator /app/bin/dex-k8s-authenticator
COPY html /app/html
COPY templates /app/templates
COPY templates /app/original-templates

WORKDIR /app

ENTRYPOINT ["/app/bin/dex-k8s-authenticator"]
CMD ["--help"]
