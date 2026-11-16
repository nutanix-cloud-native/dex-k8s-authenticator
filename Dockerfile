FROM gcr.io/distroless/static@sha256:1a64f46364620465ef9e9a8f5413757dfdf22695e050bbfd7303eeaf17a59383

COPY ./bin/linux/amd64/dex-k8s-authenticator /app/bin/dex-k8s-authenticator
COPY html /app/html
COPY templates /app/templates
COPY templates /app/original-templates

WORKDIR /app

ENTRYPOINT ["/app/bin/dex-k8s-authenticator"]
CMD ["--help"]
