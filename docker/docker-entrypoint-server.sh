#!/bin/sh

if [ -z "${CONFIG_DIR}" ]; then
    CONFIG_DIR=/etc/certsync
fi

if [ ! -d "${CONFIG_DIR}" ]; then
    mkdir "${CONFIG_DIR}"
fi

if [ $# = 1 ] && [ "$1" = 'certsync-server' ]; then
    if [ ! -f "${CONFIG_DIR}/server.yml" ]; then
        certsync-server -g -c "${CONFIG_DIR}/server.yml"
    else
        certsync-server -c "${CONFIG_DIR}/server.yml"
    fi
else
    exec "$@"
fi
