# Install

## Server Install

### Load Image(Optional)

```
docker load < path_to_certsync-server_image.tar
```

### Compose & Config Install

```
WORKDIR=$(pwd)

mkdir -p ${WORKDIR}/certsync/server/conf ${WORKDIR}/certsync/server/data

cp path_to_certsync-server_docker-compose.yml ${WORKDIR}/certsync/docker-compose.yml

cp path_to_certsync-server.yml ${WORKDIR}/certsync/server/conf/server.yml
```

### Startup

```
WORKDIR=$(pwd)

cd ${WORKDIR}/certsync

docker compose up -d
```

## Client Install

### Load Image(Optional)

```
docker load < path_to_certsync-client_image.tar
```

### Compose & Config & PostScript Install

```
WORKDIR=$(pwd)

mkdir -p ${WORKDIR}/certsync/client/conf ${WORKDIR}/certsync/client/script ${WORKDIR}/certsync/client/data

cp path_to_certsync-client_docker-compose.yml ${WORKDIR}/certsync/docker-compose.yml

cp path_to_certsync-client.yml ${WORKDIR}/certsync/client/conf/client.yml

cp path_to_certsync-postscript.sh ${WORKDIR}/certsync/client/script/certsync-postscript.sh

chmod +x ${WORKDIR}/certsync/client/script/certsync-postscript.sh
```

### Startup

```
WORKDIR=$(pwd)

cd ${WORKDIR}/certsync

docker compose up -d
```