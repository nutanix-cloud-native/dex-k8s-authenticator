FROM gcr.io/distroless/static@sha256:15f8637509792e05b97d83451b0b6ccfa4abd548f997e406ac1e72c87fe1fa78

COPY ./bin/linux/amd64/dex-k8s-authenticator /app/bin/dex-k8s-authenticator
COPY html /app/html
COPY templates /app/templates
COPY templates /app/original-templates

WORKDIR /app

ENTRYPOINT ["/app/bin/dex-k8s-authenticator"]
CMD ["--help"]
