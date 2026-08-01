#!/bin/bash
set -euo pipefail

CERT_LIVE_DIR=/share/CACHEDEV1_DATA/apps/certsync/live
STUNNEL_DIR=/etc/stunnel
STUNNEL_CONFIG_DIR=/etc/config/stunnel

# 0. Verify source files exist — bail before writing anything to avoid a
#    half-updated state on the NAS.
for f in fullchain.pem privkey.pem cert.pem chain.pem; do
  [ -f "${CERT_LIVE_DIR}/${f}" ] || { echo "qnap-certsync-post: missing ${CERT_LIVE_DIR}/${f}" >&2; exit 1; }
done

# 1. Write the `.def` source files. On QTS 5.x these are the persistent
#    "source of truth" — `stunnel.sh generate_cert_key` rebuilds the active
#    backup.key/backup.cert/stunnel.pem from them on boot/restart. Writing the
#    active files directly is the classic footgun: they get overwritten from
#    the .def on next generate_cert_key.
#
#    uca.pem is the exception: the Apache reverse proxy reads it directly from
#    /etc/config/stunnel/ (per STUNNEL_UCA in stunnel.sh; no .def, not rebuilt
#    by generate_cert_key), so we write the active file there.
cat "${CERT_LIVE_DIR}/cert.pem" "${CERT_LIVE_DIR}/chain.pem" > "${STUNNEL_DIR}/backup.cert.def"
cp    "${CERT_LIVE_DIR}/privkey.pem"                  "${STUNNEL_DIR}/backup.key.def"
[ -d "${STUNNEL_CONFIG_DIR}" ] || mkdir -p "${STUNNEL_CONFIG_DIR}"
cp    "${CERT_LIVE_DIR}/chain.pem"                    "${STUNNEL_CONFIG_DIR}/uca.pem"

# 2. Let QTS rebuild the active cert/key files (backup.key, backup.cert,
#    stunnel.pem) from the .def files with chmod 600. Caveat: the whole
#    generate_cert_key body is gated on `[ ! -f /etc/IS_G ]`, so it is a no-op
#    on IS_G (enterprise) models. Consumer/SMB boxes like the TS428 are fine.
/etc/init.d/stunnel.sh generate_cert_key

# 3. Reload the HTTPS endpoint so the new cert is live. On QTS 5.x port 443 is
#    served by apache_proxys; reload_apache does a graceful `apache -k graceful`
#    with no client-facing outage. Fall back to a full restart for older or
#    apache-less setups that run stunnel instead.
/etc/init.d/stunnel.sh reload_apache 2>/dev/null || /etc/init.d/stunnel.sh restart

echo "qnap-certsync-post: certificate deployed"
