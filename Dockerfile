# syntax=docker/dockerfile:1.7

# Pinned versions for reproducible builds. Update intentionally.
ARG CADDY_VERSION=2.10.2
# The TransIP module currently has no tagged releases; pin to a commit.
ARG TRANSIP_MODULE_VERSION=55a5d2e

FROM caddy:${CADDY_VERSION}-builder AS builder

ARG TRANSIP_MODULE_VERSION

# Build a static Caddy binary with the TransIP DNS module.
ENV CGO_ENABLED=0
RUN xcaddy build \
    --with github.com/caddy-dns/transip@${TRANSIP_MODULE_VERSION} \
    --output /usr/bin/caddy

# Build a tiny entrypoint that supports TRANSIP_PRIVATE_KEY or TRANSIP_PRIVATE_KEY__FILE.
COPY entrypoint /src/entrypoint
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /usr/bin/entrypoint /src/entrypoint/main.go

# Prepare runtime directories and default config.
RUN mkdir -p /etc/caddy /config /data
COPY Caddyfile /etc/caddy/Caddyfile

# Minimal runtime image.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
COPY --from=builder /usr/bin/entrypoint /usr/bin/entrypoint
COPY --from=builder /etc/caddy /etc/caddy
COPY --from=builder /config /config
COPY --from=builder /data /data

USER 65532:65532
EXPOSE 80 443
VOLUME ["/data", "/config"]

ENTRYPOINT ["/usr/bin/entrypoint"]
CMD ["run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]
