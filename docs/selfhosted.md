# Self-Hosting ReMemory

Run ReMemory as a web app on your own server — create bundles, store encrypted archives, and recover, all from a browser.

## Getting started

### Docker

A pre-built image is published to GitHub Container Registry on every release.

```bash
docker run -d \
  --name rememory \
  -p 8080:8080 \
  -v rememory-data:/data \
  ghcr.io/eljojo/rememory:latest
```

Visit `http://localhost:8080` to set up. The first page asks you to choose an admin password for deleting bundles.

To pin a specific version:

```bash
docker run -d \
  --name rememory \
  -p 8080:8080 \
  -v rememory-data:/data \
  ghcr.io/eljojo/rememory:v0.0.16
```

**Docker Compose:**

```yaml
services:
  rememory:
    image: ghcr.io/eljojo/rememory:latest
    ports:
      - "8080:8080"
    volumes:
      - rememory-data:/data
    restart: unless-stopped

volumes:
  rememory-data:
```

The container is a single static binary with no dependencies. Data lives in `/data` — mount a volume there to persist across restarts.

### Without Docker

If you have the CLI installed:

```bash
rememory serve
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--port, -p` | `8080` | Port to listen on |
| `--host` | `127.0.0.1` | Host to bind to |
| `--data, -d` | `./rememory-data` | Data directory for bundles and config |
| `--max-manifest-size` | `50MB` | Maximum MANIFEST.age size (e.g. `50MB`, `1GB`) |
| `--no-timelock` | false | Omit time-lock support |

## Deployment

### Reverse proxy

Put the server behind a reverse proxy with TLS.

**Caddy:**
```
rememory.example.com {
    reverse_proxy localhost:8080
}
```

**nginx:**
```nginx
server {
    listen 443 ssl;
    server_name rememory.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        client_max_body_size 100M;
    }
}
```

### Authentication

The admin password only protects bundle deletion. For access control, use an auth proxy:

- [Authelia](https://www.authelia.com/)
- [Cloudflare Access](https://www.cloudflare.com/products/zero-trust/access/)
- [Pocket ID](https://github.com/pocket-id/pocket-id)
- [OAuth2 Proxy](https://oauth2-proxy.github.io/oauth2-proxy/)

## Security

- The server stores only encrypted archives (MANIFEST.age). Without enough shares, the archive is useless.
- Shares are never sent to the server. They stay in each friend's bundle.
- The admin password uses age's scrypt-based passphrase encryption. Choose a strong one.
- Put the server behind HTTPS and authentication appropriate for your threat model.

Friends still get self-contained offline bundles. The server is a convenience — if it goes away, they can recover without it.

## Data directory

The data directory contains:

```
rememory-data/
  admin.age               # Admin password (age-encrypted)
  bundles/
    <uuid>/
      meta.json           # Non-secret metadata
      MANIFEST.age        # Encrypted archive
```

Back up this directory to preserve your encrypted archives. The admin.age file can be recreated by setting a new password (you'd lose the ability to delete existing bundles with the old password, but the bundles themselves are unaffected).
