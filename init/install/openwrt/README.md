# Install

## Server Install

### Binary Install

```
cp path_to_certsync-server_binary /usr/bin/certsync-server
chmod +x /usr/bin/certsync-server
```

### Config Install

```
mkdir -p /etc/certsync /var/lib/certsync

cd /etc/certsync

cp path_to_certsync-server.yml /etc/certsync/server.yml
```

### AutoStartup

```
cp path_to_certsync-server_init.d_file /etc/init.d/certsync-server

chmod +x /etc/init.d/certsync-server

/etc/init.d/certsync-server enable
/etc/init.d/certsync-server start
```

## Client Install

### Binary Install

```
cp path_to_certsync-client_binary /usr/bin/certsync-client

chmod +x /usr/bin/certsync-client
```

### Config & PostScript Install

```
mkdir -p /etc/certsync/live

cd /etc/certsync

cp path_to_certsync-client.yml /etc/certsync/client.yml
cp path_to_certsync-postscript.sh /usr/bin/certsync-postscript.sh

chmod +x /usr/bin/certsync-postscript.sh
```

### AutoStartup

```
cp path_to_certsync-client_init.d_file /etc/init.d/certsync-client

chmod +x /etc/init.d/certsync-client

/etc/init.d/certsync-client enable
/etc/init.d/certsync-client start
```