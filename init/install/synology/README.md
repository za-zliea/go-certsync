# Install

## Server Install

### Binary Install

```
cp path_to_certsync-server_binary /usr/local/bin/certsync-server
chmod +x /usr/local/bin/certsync-server
```

### Config Install

```
mkdir -p /etc/certsync /var/lib/certsync

cd /etc/certsync

cp path_to_certsync-server.yml /etc/certsync/server.yml
```

### AutoStartup

```
cp path_to_certsync-server_systemd.service /usr/lib/systemd/system/certsync-server.service

synosystemctl daemon-reload
synosystemctl enable certsync-server.service
synosystemctl start certsync-server.service
```

## Client Install

### Binary Install

```
cp path_to_certsync-client_binary /usr/local/bin/certsync-client

chmod +x /usr/local/bin/certsync-client
```

### Config & PostScript Install

```
mkdir -p /etc/certsync/live

cd /etc/certsync

cp path_to_certsync-client.yml /etc/certsync/client.yml
cp path_to_certsync-postscript.sh /usr/local/bin/certsync-postscript.sh

chmod +x /usr/local/bin/certsync-postscript.sh
```

### AutoStartup

```
cp path_to_certsync-client_systemd.service /usr/lib/systemd/system/certsync-client.service

synosystemctl daemon-reload
synosystemctl enable certsync-client.service
synosystemctl start certsync-client.service
```