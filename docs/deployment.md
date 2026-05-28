# Deployment

## Infrastructure

The API runs on a NixOS server alongside other apps (e.g. familyArchive).
All apps share the same infrastructure stack:

- **Reverse proxy:** Traefik v3 with Let's Encrypt + Cloudflare DNS challenge
- **Database:** PostgreSQL with PgBouncer connection pooling
- **Containers:** Docker via NixOS `oci-containers`
- **Secrets:** sops-nix for encrypted secret management
- **Networking:** Docker bridge networks (`proxy-net` for Traefik, `postgres-net` for DB)

## Adding the App to the Server

Follow the checklist in `nixos-server/apps/_template.nix`:

1. Copy `_template.nix` to `apps/home-cooking.nix`
2. Set `appName = "home-cooking"`, fill in domain and other variables
3. Add database entry to `modules/postgresql.nix`:
   ```nix
   { name = "home-cooking"; user = "home_cooking"; dbName = "home_cooking"; }
   ```
4. Add secrets to `secrets/secrets.yaml` and declare in `modules/secrets.nix`
5. Import in `configuration.nix`
6. Deploy: `nixos-rebuild switch --flake ~/nixos-server#server`
7. Set database password:
   ```
   sudo -u postgres psql -c "ALTER USER home_cooking WITH PASSWORD '...';"
   ```
8. Restart PgBouncer:
   ```
   sudo systemctl restart pgbouncer-auth && sudo systemctl restart pgbouncer
   ```

## CI/CD Pipeline

The release workflow (`.github/workflows/release.yml`) triggers on GitHub
release publish:

1. Builds Docker image from `api/` directory
2. Pushes to GHCR with semver + sha tags
3. SSHs into server, runs `deploy.sh home-cooking <digest>`
4. `deploy.sh` updates the image SHA in the NixOS app config and runs
   `nixos-rebuild switch`

## Local Development

```bash
cd api/

# Start Postgres
docker compose up postgres -d

# Run the server (auto-migrates)
make run

# Or run everything in containers
make docker-up
```

## Environment Variables

| Variable       | Description                    | Default (dev)                                              |
|----------------|--------------------------------|------------------------------------------------------------|
| `PORT`         | Server listen port             | `8080`                                                     |
| `DATABASE_URL` | PostgreSQL connection string   | `postgres://homecooking:homecooking@localhost:5432/homecooking?sslmode=disable` |

Production secrets (DATABASE_URL with real credentials, etc.) are injected
via sops-nix environment files — see the NixOS app config.
