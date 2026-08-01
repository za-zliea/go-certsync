#!/bin/bash

set -euo pipefail

CERT_LIVE_DIR=/etc/certsync/live

CERT_DEST_DIR=/etc/nginx/ssl

for f in fullchain.pem privkey.pem cert.pem chain.pem; do
  [ -f "${CERT_LIVE_DIR}/${f}" ] || { echo "qnap-certsync-post: missing ${CERT_LIVE_DIR}/${f}" >&2; exit 1; }
done

# Copy the certs to the destination directory
cp "${CERT_LIVE_DIR}/fullchain.pem" "${CERT_DEST_DIR}/fullchain.pem"
cp "${CERT_LIVE_DIR}/privkey.pem" "${CERT_DEST_DIR}/privkey.pem"
cp "${CERT_LIVE_DIR}/chain.pem" "${CERT_DEST_DIR}/chain.pem"
cp "${CERT_LIVE_DIR}/cert.pem" "${CERT_DEST_DIR}/cert.pem"

nginx -t
nginx -s reload