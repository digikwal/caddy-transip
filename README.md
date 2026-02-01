# Caddy + TransIP DNS-01 (custom build)

This repository builds a **custom Caddy v2** image with the `dns.providers.transip` module included. The module source is `github.com/caddy-dns/transip` (the Caddy DNS modules org), pinned to `v2.0.1` for reproducibility.

## How TransIP DNS-01 works

For DNS-01 challenges, Caddy asks the ACME CA for a token, then creates a TXT record under `_acme-challenge.<domain>` via the TransIP API. Once the CA can resolve that TXT record, it validates ownership and issues the certificate. This is required for wildcard certificates like `*.example.nl`.

## Environment variables

The Caddyfile reads credentials from environment variables (the entrypoint supports both raw keys and Docker-style `__FILE` secrets):

- `TRANSIP_ACCOUNT_NAME`: your TransIP login / account name.
- `TRANSIP_PRIVATE_KEY`: either raw private key content or a path to a key file.
- `TRANSIP_PRIVATE_KEY__FILE`: path to a file containing the private key.

Tip: if you keep the key in a file, you can pass the path with `TRANSIP_PRIVATE_KEY__FILE`, or export the file contents before starting the container:

```
export TRANSIP_PRIVATE_KEY="$(cat /path/to/transip.key)"
```

The entrypoint normalizes credentials into `TRANSIP_PRIVATE_KEY` for Caddy:

- If `TRANSIP_PRIVATE_KEY__FILE` is set, it points Caddy to that file path.
- If `TRANSIP_PRIVATE_KEY` points to an existing file, that path is used directly.
- If `TRANSIP_PRIVATE_KEY` contains raw key content, it writes `/tmp/transip.key` (mode 0600) and points Caddy to that file.

## Default Caddyfile example

The included `Caddyfile` is a minimal wildcard example:

```
*.example.nl {
  tls {
    dns transip {env.TRANSIP_ACCOUNT_NAME} {env.TRANSIP_PRIVATE_KEY}
  }
  respond "ok"
}
```

Mount your own `Caddyfile` in production.

## Docker usage

Build locally:

```
docker build -t caddy-transip:local .
```

Run with `docker run`:

```
docker run --rm -p 80:80 -p 443:443 \
  -e TRANSIP_ACCOUNT_NAME="your-login" \
  -e TRANSIP_PRIVATE_KEY="your-private-key" \
  -v $(pwd)/Caddyfile:/etc/caddy/Caddyfile:ro \
  -v caddy_data:/data \
  -v caddy_config:/config \
  ghcr.io/<owner>/<repo>:latest
```

Or use a mounted key file:

```
docker run --rm -p 80:80 -p 443:443 \
  -e TRANSIP_ACCOUNT_NAME="your-login" \
  -e TRANSIP_PRIVATE_KEY__FILE="/run/secrets/transip.key" \
  -v /path/to/transip.key:/run/secrets/transip.key:ro \
  -v $(pwd)/Caddyfile:/etc/caddy/Caddyfile:ro \
  -v caddy_data:/data \
  -v caddy_config:/config \
  ghcr.io/<owner>/<repo>:latest
```

Run with `docker-compose`:

```
services:
  caddy:
    image: ghcr.io/<owner>/<repo>:latest
    ports:
      - "80:80"
      - "443:443"
    environment:
      TRANSIP_ACCOUNT_NAME: "your-login"
      TRANSIP_PRIVATE_KEY: "your-private-key"
      # Or: TRANSIP_PRIVATE_KEY__FILE: "/run/secrets/transip.key"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config

volumes:
  caddy_data:
  caddy_config:
```

## CI/CD and releases

- GitHub Actions builds, tests (`caddy version` + module presence), and pushes the image to GHCR.
- Releases are driven by **semantic-release** using **Conventional Commits**.
- Docker tags are created automatically: `latest`, `vX`, `vX.Y`, and `vX.Y.Z`.
- Images are signed with **Cosign** and an **SBOM** is generated with **Syft**.

## Notes

- This is a **custom Caddy build** (not the upstream image) to include `dns.providers.transip`.
- Caddy runs with: `caddy run --config /etc/caddy/Caddyfile --adapter caddyfile`.
