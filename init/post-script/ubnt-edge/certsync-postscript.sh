#!/bin/bash

CERT_LIVE_DIR=/etc/certsync/live

cat ${CERT_LIVE_DIR}/privkey.pem ${CERT_LIVE_DIR}/fullchain.pem > /config/auth/certificates/server.pem

systemctl restart lighttpd