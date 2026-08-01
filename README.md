# go-certsync

A certificate synchronization service for SSL/TLS certificates [GitHub](https://github.com/za-zliea/go-certsync)

## Feature

### Server

- Base on [atreugo](https://github.com/savsgio/atreugo) Web Server Framework.
- Automatic certificate renewal using ACME (Let's Encrypt).
- Certificate distribution via HTTP API.
- Manual certificate upload via web page or API.

### Client

- Periodically check and sync certificates from server.
- Support custom update commands after certificate sync.

### Support DNS Provider

- Tencent Cloud DNS
- Aliyun DNS
- GoDaddy
- Google Cloud DNS
- Cloudflare

## Build

### Prerequirements

- golang >= 1.25
- make

### Without Docker

```shell
make all
```

### With Docker

```shell
make image VERSION=[GIT TAG]
make image-alpine VERSION=[GIT TAG]
```

## Usage

### Server Usage

```
Usage:
  server startup:
    certsync-server [-c config file]
  server startup in background:
    nohup certsync-server [-c config file] &
  generate demo config file:
    certsync-server -g [-c config file]
  print usage:
    certsync-server -h
  print version and exit:
    certsync-server -v
Options:
  -c string
    	config file path, default server.conf (default "server.conf")
  -g	generate config, default server.conf
  -h	print usage
  -v	print version and exit
```

### Client Usage

```shell
Usage:
  client startup:
    certsync-client [-c config file]
  client startup in background:
    nohup certsync-client [-c config file] &
  generate demo config file:
    certsync-client -g [-c config file]
  print usage:
    certsync-client -h
  print version and exit:
    certsync-client -v
Options:
  -c string
    	config file path, default client.conf (default "client.conf")
  -g	generate config, default client.conf
  -h	print usage
  -v	print version and exit
```

## Docker

### Server

[Docker Hub](https://hub.docker.com/r/zliea/certsync-server)

```shell
docker run -d -p 8080:8080 --name certsync-server -v ./:/etc/certsync zliea/certsync-server:latest
```

### Client

[Docker Hub](https://hub.docker.com/r/zliea/certsync-client)

```shell
docker run -d --name certsync-client -v ./:/etc/certsync zliea/certsync-client:latest
```

## Systemd

### Install

```shell
# Build
make build

# Install binary, systemd service and generate config (requires root)
sudo make install

# Edit config
vi /etc/certsync/server.conf
vi /etc/certsync/client.conf

# Enable and start service
sudo systemctl enable --now certsync-server
sudo systemctl enable --now certsync-client
```

### Uninstall

```shell
sudo make uninstall
```

## Config

### Server Config

```yaml
address: 0.0.0.0                        # Listen address
port: 8080                              # Listen port
token: your-server-token-abcde12345     # Client/API auth token (also HMAC key for admin JWTs)
storage: /var/lib/certsync              # Certificate storage directory
cert_check_time: "03:00:00"             # Daily certificate check time (HH:MM:SS)
dns: 8.8.8.8                            # DNS server for propagation check (optional)
admin:                                  # Admin account for the web UI (auto-generated if omitted)
  username: admin
  password: bcrypt-hashed-password      # bcrypt hash; plaintext is logged once on first-run generation
certs:
  - alias: example.com                  # Certificate alias (used in API)
    auth: your-cert-auth-token          # Certificate auth token
    email: admin@example.com            # Email for Let's Encrypt registration
    domain: example.com                 # Domain for certificate
    domain_cn:                          # Subject Alternative Names
      - example.com
      - www.example.com
    domain_check: example.com:443       # host:port used by scheduler to probe the live cert
    provider: TENCENT                   # DNS provider (TENCENT/ALIYUN/GODADDY/GOOGLE/CLOUDFLARE)
    ak: your-access-key                 # Access Key ID
    sk: your-access-key-secret          # Access Key Secret
    auto_renew: true                    # Enable automatic renewal via Let's Encrypt
    disable_cname: false                # Optional: disable CNAME support during ACME DNS-01
    upload_token: your-upload-token     # Legacy upload token (unused by /front/* JWT flow)
```

On first run, if `admin` is absent or empty the server generates a random 16-character
password, bcrypt-hashes it, persists it to the config file, and logs the plaintext once with
a warning to change it after login.

#### Certificate Modes

| `auto_renew` | Mode | Description |
|--------------|------|-------------|
| `true` | Auto Renewal | Server automatically obtains/renews certificates via Let's Encrypt ACME |
| `false` | Manual Upload | Certificates must be uploaded manually via API or web page |

### Client Config

```yaml
server: https://certsync.example.com    # Server URL
token: your-server-token-abcde12345     # Server auth token
cert_alias: example.com                 # Certificate alias
cert_auth: your-cert-auth-token         # Certificate auth token
domain_check: example.com:443           # host:port of the live cert to probe
local_check: false                      # true = use /api/{alias}/expire flow (default false)
cert_update_dir: /etc/nginx/ssl/example.com  # Directory to store certificates
cert_update_cmd: nginx -s reload        # Command to run after certificate update
interval: 86400                         # Check interval in seconds (default: 86400 = 24 hours)
```

### Domain Check Configuration

`domain_check` is a `host:port` value (port defaults to 443 if omitted), not a URL. It is
dialed over TLS to read the `NotAfter` of the certificate currently serving on your domain.
There are two check flows, selected by the client's `local_check` flag:

**`local_check: false` (default) — server-side compare via `/api/{alias}/check`:**
1. Client sends its `domain_check` to the server as a query parameter.
2. Server dials `domain_check` to get the remote expiry, reads its own local cert expiry.
3. Server returns `need_update = remote_expires_before_local`.

**`local_check: true` — client-side compare via `/api/{alias}/expire`:**
1. Client dials `domain_check` itself to get the remote expiry.
2. Client posts `remoteExpire` to `/api/{alias}/expire`; server returns its local expiry.
3. If the server has no local cert yet, this request also triggers an async renewal
   (see commit `6b2996a`), and the next client tick picks up the renewed certificate.
4. Client returns `need_update = remote_expires_before_server`.

A separate per-cert `domain_check` on the server config is used by the scheduler to decide
whether to renew when no local cert exists.

## API

All `/api/*` endpoints are used by clients and require `Authorization: {server_token}` plus
`?auth={cert_auth}`. The `/front/*` and `/pub/login` endpoints back the admin web UI and use
JWT bearer auth (`Authorization: Bearer <jwt>`).

### Check Certificate (client, `local_check: false`)

```
GET /api/{alias}/check?auth={cert_auth}&domain_check=host:443
Authorization: {server_token}
```

Response:
```json
{
  "code": 0,
  "message": "SUCCESS",
  "data": {
    "need_update": true,
    "local_expiry": "2024-01-01T00:00:00Z",
    "remote_expiry": "2024-03-01T00:00:00Z",
    "reason": "remote certificate expires before local"
  }
}
```

### Get Server Expiry (client, `local_check: true`)

```
GET /api/{alias}/expire?auth={cert_auth}&remoteExpire=RFC3339
Authorization: {server_token}
```

Returns the server's local cert expiry. If the server has no local cert and `remoteExpire`
indicates renewal is needed, the server triggers an async renewal and returns 500 for this
request — retry on the next tick.

### Download Certificate (client)

```
GET /api/{alias}/download?auth={cert_auth}
Authorization: {server_token}
```

Response: ZIP file containing `cert.pem`, `chain.pem`, `fullchain.pem`, and `privkey.pem`

### Admin Login

```
POST /pub/login
Content-Type: application/json

{"username": "admin", "password": "..."}
```

Response:
```json
{
  "code": 0,
  "message": "SUCCESS",
  "data": {"token": "<jwt>", "expires_at": 1700000000}
}
```

The JWT is HS256-signed with the server `token` as the HMAC key and is valid for 8 hours.

### List Certificates (admin)

```
GET /front/list
Authorization: Bearer <jwt>
```

Returns all configured certificates with secrets (`ak`, `sk`, `auth`, `upload_token`) stripped.

### Download / Upload via Web UI (admin)

```
GET  /front/{alias}/download          # ZIP download, same format as client API
POST /front/{alias}/upload            # multipart: fullchain + privkey (auto_renew: false only)
Authorization: Bearer <jwt>
```

## Web UI

The server binary embeds a single-page admin frontend (built into `src/frontend/dist/` and
served at `/` via `go:embed`). Open it at:

```
http://your-server:8080/
```

Log in with the admin account from the server config. On first run, if no admin is configured
the server generates a random password, hashes it, persists it to the config file, and logs
the plaintext once — change it from the UI after first login.

From the UI you can:
- View all configured certificates (without exposing DNS provider secrets)
- Download any certificate as a ZIP
- Upload `fullchain.pem` + `privkey.pem` for certificates with `auto_renew: false`

Manual upload is rejected for certificates with `auto_renew: true`.
