#!/bin/sh

if [ -z "${CONFIG_DIR}" ]; then
    CONFIG_DIR=/etc/certsync
fi

if [ ! -d "${CONFIG_DIR}" ]; then
    mkdir "${CONFIG_DIR}"
fi

if [ $# = 1 ] && [ "$1" = 'certsync-client' ]; then
    if [ ! -f "${CONFIG_DIR}/client.yml" ]; then
        certsync-client -g -c "${CONFIG_DIR}/client.yml"
    else
        certsync-client -c "${CONFIG_DIR}/client.yml"
    fi
else
    exec "$@"
fi
