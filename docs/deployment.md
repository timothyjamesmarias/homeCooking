# Deployment

## Infrastructure

The API runs on a NixOS server alongside other apps (e.g. familyArchive).
All apps share the same infrastructure stack:

- **Reverse proxy:** Traefik v3 with Let's Encrypt + Cloudflare DNS challenge
- **Database:** PostgreSQL with PgBouncer connection pooling
- **Containers:** Docker via NixOS `oci-containers`
- **Secrets:** sops-nix for encrypted secret management
- **Networking:** Docker bridge networks (`proxy-net` for Traefik, `postgres-net` for DB)

## Server Permissions Model

The `deploy` user has SSH key auth only (no password). Sudo is scoped:

- **As root (NOPASSWD):** `nixos-rebuild`, `systemctl`, `nix-collect-garbage`
- **As postgres (NOPASSWD):** all commands (postgres user is the security boundary)
- **Root password:** managed via sops with `neededForUsers = true`, available via `su -`

`users.mutableUsers = false` is set so that passwords are applied declaratively
on every rebuild. Without this, NixOS only sets passwords on first boot and
ignores changes in subsequent rebuilds.

## Adding a New App to the Server

1. Copy `nixos-server/apps/_template.nix` to `apps/<app-name>.nix`
2. Set `appName`, `domain`, `ghOrg`, `appPort`
3. Add database entry to `modules/postgresql.nix`:
   ```nix
   { name = "my-app"; user = "my_app"; dbName = "my_app"; }
   ```
4. Add secrets to `secrets/secrets.yaml` via sops:
   ```bash
   sops secrets/secrets.yaml
   # add: my-app/database-password: <password>
   ```
5. Declare secrets in `modules/secrets.nix` with `restartUnits`:
   ```nix
   "my-app/database-password" = {
     owner = "root";
     restartUnits = [ "db-passwords.service" "my-app-env.service" "docker-my-app.service" ];
   };
   ```
6. Import in `configuration.nix`
7. Add DNS record in Cloudflare pointing to the server
8. Push config, pull on server, rebuild:
   ```
   sudo nixos-rebuild switch --flake ~/nixos-server#server
   ```

Database passwords are fully automated:
- `db-passwords` service reads passwords from sops and runs `ALTER USER` on every rebuild
- `pgbouncer-auth` regenerates the auth file after passwords are set
- App env services generate `DATABASE_URL` from sops
- sops `restartUnits` ensures all of the above re-run when a secret changes

## CI/CD Pipeline

The release workflow (`.github/workflows/release.yml`) triggers on tag push
matching `api/v*`:

```bash
git tag api/v1.0.0
git push origin api/v1.0.0
```

Pipeline steps:
1. Builds Docker image from `api/` directory
2. Pushes to GHCR with version + sha tags
3. SSHs into server, runs `deploy.sh home-cooking <digest>`
4. `deploy.sh` updates the image SHA in the NixOS app config and runs
   `nixos-rebuild switch`

### Required GitHub Secrets

- `SERVER_HOST` — server IP or hostname
- `SERVER_USER` — `deploy`
- `SERVER_SSH_KEY` — deploy user's SSH private key (must include BEGIN/END
  headers and preserve newlines when pasting into GitHub)

## Troubleshooting

### Container crash-looping (start-limit-hit)

If a container fails repeatedly, systemd stops trying. After fixing the
underlying issue, you must reset the failure counter:

```bash
systemctl reset-failed docker-<app-name>
sleep 1
systemctl start docker-<app-name>
```

### SASL authentication failed

The password in the container's DATABASE_URL doesn't match what Postgres or
PgBouncer expects. With the automated password flow, this should not happen
on a clean rebuild. If it does:

- Verify the env file: `cat /run/<app-name>/env`
- Test directly against Postgres (bypassing PgBouncer):
  ```bash
  PGPASSWORD='...' psql -h 127.0.0.1 -p 5432 -U <user> -d <db> -c "SELECT 1"
  ```
- Test through PgBouncer:
  ```bash
  PGPASSWORD='...' psql -h 127.0.0.1 -p 6432 -U <user> -d <db> -c "SELECT 1"
  ```
- If Postgres works but PgBouncer doesn't, check that `pgbouncer-auth`
  ran after `db-passwords`: `systemctl status pgbouncer-auth`

### Migration errors with empty migrations directory

golang-migrate's file source driver chokes on non-SQL files (e.g. `.gitkeep`).
The Dockerfile strips non-SQL files from the migrations directory. The server
also skips migrations entirely if no `.sql` files are found.

### Deploy script updated SHA but container still runs old image

The deploy script modifies the nix file on the server via `sed` and rebuilds.
If the container was already crash-looping, the rebuild may not reset the
failure counter. Use `systemctl reset-failed` as described above.

### sops passwords not applying

- `users.mutableUsers` must be `false` or NixOS won't overwrite existing
  password entries in `/etc/shadow`
- Secrets with `neededForUsers = true` go to `/run/secrets-for-users/`,
  not `/run/secrets/`
- Verify sops decryption: `sops -d secrets/secrets.yaml | grep <key>`

## Deploy Script

The deploy script (`nixos-server/scripts/deploy.sh`) handles:
- Resetting any previous container failure state
- Updating the image SHA and rebuilding
- Waiting for the container to start (30s timeout)
- Hitting the `/health` endpoint to verify the app is routable
- Printing container logs on failure for CI visibility

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
