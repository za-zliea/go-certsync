# go-certsync

A certificate synchronization service for SSL/TLS certificates [GitHub](https://github.com/za-zliea/go-certsync)

## Feature

### Server

- Base on [atreugo](https://github.com/savsgio/atreugo) Web Server Framework.
- Automatic certificate renewal using ACME (Let's Encrypt).
- Certificate distribution via HTTP API.

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
Options:
  -c string
    	config file path, default server.conf (default "server.conf")
  -g	generate config, default server.conf
  -h	print usage
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
Options:
  -c string
    	config file path, default client.conf (default "client.conf")
  -g	generate config, default client.conf
  -h	print usage
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

# Install binary and systemd service (requires root)
sudo make install

# Create config directory
sudo mkdir -p /etc/certsync

# Generate config
certsync-server -g -c /etc/certsync/server.conf
certsync-client -g -c /etc/certsync/client.conf

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
token: your-server-token-abcde12345     # Client and server auth token
storage: /var/lib/certsync              # Certificate storage directory
cert_check_time: "030000"               # Daily certificate check time (HHMMSS)
certs:
  - alias: example.com                  # Certificate alias (used in API)
    auth: your-cert-auth-token          # Certificate auth token
    domain: example.com                 # Domain for certificate
    domain_check: https://example.com   # URL to check current certificate
    provider: TENCENT                   # DNS provider (TENCENT/ALIYUN/GODADDY/GOOGLE/CLOUDFLARE)
    ak: your-access-key                 # Access Key ID
    sk: your-access-key-secret          # Access Key Secret
```

### Client Config

```yaml
server: https://certsync.example.com    # Server URL
token: your-server-token-abcde12345     # Server auth token
cert_alias: example.com                 # Certificate alias
cert_auth: your-cert-auth-token         # Certificate auth token
cert_update_dir: /etc/nginx/ssl/example.com  # Directory to store certificates
cert_update_cmd: nginx -s reload        # Command to run after certificate update
interval: 86400                         # Check interval in seconds (default: 86400 = 24 hours)
```

## API

### Check Certificate

```
GET /api/{alias}/check?auth={cert_auth}
Authorization: {server_token}
```

Response:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "need_update": true,
    "local_expiry": "2024-01-01T00:00:00Z",
    "remote_expiry": "2024-03-01T00:00:00Z",
    "reason": "remote certificate expires before local"
  }
}
```

### Download Certificate

```
GET /api/{alias}/download?auth={cert_auth}
Authorization: {server_token}
```

Response: ZIP file containing `cert.pem`, `chain.pem`, `fullchain.pem`, and `privkey.pem`
