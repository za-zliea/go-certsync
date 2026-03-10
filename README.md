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
token: your-server-token-abcde12345     # Client and server auth token
storage: /var/lib/certsync              # Certificate storage directory
cert_check_time: "03:00:00"             # Daily certificate check time (HH:MM:SS)
dns: 8.8.8.8                            # DNS server for propagation check (optional)
certs:
  - alias: example.com                  # Certificate alias (used in API)
    auth: your-cert-auth-token          # Certificate auth token
    email: admin@example.com            # Email for Let's Encrypt registration
    domain: example.com                 # Domain for certificate
    domain_cn:                          # Subject Alternative Names
      - example.com
      - www.example.com
    provider: TENCENT                   # DNS provider (TENCENT/ALIYUN/GODADDY/GOOGLE/CLOUDFLARE)
    ak: your-access-key                 # Access Key ID
    sk: your-access-key-secret          # Access Key Secret
    auto_renew: true                    # Enable automatic renewal via Let's Encrypt
    upload_token: your-upload-token     # Token for manual upload (used when auto_renew: false)
```

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
domain_check: https://example.com       # URL to check remote certificate expiry
cert_update_dir: /etc/nginx/ssl/example.com  # Directory to store certificates
cert_update_cmd: nginx -s reload        # Command to run after certificate update
interval: 86400                         # Check interval in seconds (default: 86400 = 24 hours)
```

### Domain Check Configuration

The `domain_check` field in the client config specifies the URL used to check the remote certificate's expiry date. This URL is passed to the server during the check request, allowing the server to compare the remote certificate (the one currently serving on your domain) with the local certificate stored on the server.

**How it works:**
1. Client sends a check request with `domain_check` URL to the server
2. Server fetches the remote certificate from the `domain_check` URL
3. Server compares remote certificate expiry with local certificate expiry
4. If remote expires before local, client needs to update (newer certificate available on server)

## API

### Check Certificate

```
GET /api/{alias}/check?auth={cert_auth}&domain_check={domain_url}
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

### Verify Upload Token

```
GET /api/{alias}/upload_verify?upload_token={upload_token}
```

Response:
```json
{
  "code": 0,
  "message": "upload token is valid"
}
```

### Upload Certificate

```
POST /api/{alias}/upload?upload_token={upload_token}
Content-Type: multipart/form-data

fullchain: <fullchain.pem file>
privkey: <privkey.pem file>
```

Response:
```json
{
  "code": 0,
  "message": "certificate uploaded successfully"
}
```

## Web Upload Page

Access the web-based certificate upload page at:

```
http://your-server:8080/h5/upload
```

This page provides a user-friendly interface for manually uploading certificates:
1. Enter the certificate alias and upload token
2. After token verification, select the fullchain and private key files
3. Click upload to submit the certificate

Note: Manual upload is only available when `auto_renew: false` is set for the certificate.
