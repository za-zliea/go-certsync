# Install

## Enable autorun.sh

```
Console -> Hardware -> General -> Enable autorun.sh handling
```

## Server Install

### Binary Install

```
sudo -i

mkdir -p /share/CACHEDEV1_DATA/apps/certsync/bin /share/CACHEDEV1_DATA/apps/certsync/log

cp path_to_certsync-server_binary /share/CACHEDEV1_DATA/apps/certsync/bin/certsync-server

chmod +x /share/CACHEDEV1_DATA/apps/certsync/bin/certsync-server
```

### Config Install

```
mkdir -p /share/CACHEDEV1_DATA/apps/certsync/conf /share/CACHEDEV1_DATA/apps/certsync/data

cd /share/CACHEDEV1_DATA/apps/certsync/conf

cp path_to_certsync-server.yml /share/CACHEDEV1_DATA/apps/certsync/conf/server.yml
```

### AutoStartup

```
/share/CACHEDEV1_DATA/apps/myapp/myapp

touch /tmp/config/autorun.sh
chmod +x /tmp/config/autorun.sh

vi /tmp/config/autorun.sh

### autorun ###
#!/bin/sh

sleep 60

nohup /share/CACHEDEV1_DATA/apps/certsync/bin/certsync-server -c /share/CACHEDEV1_DATA/apps/certsync/conf/server.yml > /share/CACHEDEV1_DATA/apps/certsync/log/server.log 2>&1 &

### autorun ###

umount /tmp/config
```

## Client Install

### Binary Install

```
sudo -i

mkdir -p /share/CACHEDEV1_DATA/apps/certsync/bin /share/CACHEDEV1_DATA/apps/certsync/log

cp path_to_certsync-client_binary /share/CACHEDEV1_DATA/apps/certsync/bin/certsync-client

chmod +x /share/CACHEDEV1_DATA/apps/certsync/bin/certsync-client
```

### Config & PostScript Install

```
mkdir -p /share/CACHEDEV1_DATA/apps/certsync/conf /share/CACHEDEV1_DATA/apps/certsync/live

cp path_to_certsync-client.yml /share/CACHEDEV1_DATA/apps/certsync/conf/client.yml
cp path_to_certsync-postscript.sh /share/CACHEDEV1_DATA/apps/certsync/bin/certsync-postscript.sh

chmod +x /share/CACHEDEV1_DATA/apps/certsync/bin/certsync-postscript.sh
```

### AutoStartup

```
touch /tmp/config/autorun.sh
chmod +x /tmp/config/autorun.sh

vi /tmp/config/autorun.sh

### autorun ###
#!/bin/sh

sleep 60

nohup /share/CACHEDEV1_DATA/apps/certsync/bin/certsync-client -c /share/CACHEDEV1_DATA/apps/certsync/conf/client.yml > /share/CACHEDEV1_DATA/apps/certsync/log/client.log 2>&1 &

### autorun ###

umount /tmp/config
```