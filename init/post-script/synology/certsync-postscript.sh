#!/bin/bash

CERT_LIVE_DIR=/etc/certsync/live

for f in fullchain.pem privkey.pem cert.pem chain.pem; do
  [ -f "${CERT_LIVE_DIR}/${f}" ] || { echo "qnap-certsync-post: missing ${CERT_LIVE_DIR}/${f}" >&2; exit 1; }
done

# Look up the default cert's id and desc via the DSM API — this is the source
# of truth (more reliable than reading _archive/DEFAULT and _archive/INFO).
SYNOLOGY_CERT_LIST_JSON=$(/usr/syno/bin/synowebapi --exec api=SYNO.Core.Certificate.CRT method=list version=1)
SYNOLOGY_CERTID=$(echo "${SYNOLOGY_CERT_LIST_JSON}" | jq -r '.data.certificates[] | select(.is_default == true or .is_default == "true") | .id' | head -n1)

if [ -z "${SYNOLOGY_CERTID}" ]; then
  echo "certsync-post: no default cert found in SYNO.Core.Certificate.CRT::list output" >&2
  exit 1
fi

/usr/syno/bin/synowebapi --exec api=SYNO.Core.Certificate method=import version=1 key_tmp=${CERT_LIVE_DIR}/privkey.pem cert_tmp=${CERT_LIVE_DIR}/cert.pem inter_cert_tmp=${CERT_LIVE_DIR}/chain.pem id=${SYNOLOGY_CERTID} desc=
