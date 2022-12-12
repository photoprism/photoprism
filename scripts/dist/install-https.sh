#!/usr/bin/env bash

# Generates local HTTPS keys and certificates on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-https.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Abort if not executed as root..
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

# shellcheck disable=SC2164
CONF_PATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )/openssl"
CERTS_PATH="/etc/ssl/certs"
KEY_PATH="/etc/ssl/private"

# Check if keys and certificates already exist.
if [ -f "$CERTS_PATH/photoprism.local.issuer.crt" ] && [ -f "$KEY_PATH/photoprism.local.pfx" ] && [ -f "$KEY_PATH/photoprism.me.pfx" ]; then
    echo "Keys and certificates for photoprism.local already exist in ${KEY_PATH} and ${CERTS_PATH}."
    exit 0
fi

echo "Creating local HTTPS keys and certificates in ${KEY_PATH} and ${CERTS_PATH}."

mkdir -p "${CERTS_PATH}" "${KEY_PATH}"
groupadd -f -r -g 116 ssl-cert 1>&2

# Generate issuer (CA) certificate.

echo "Generating issuer (CA) certificate..."

openssl genrsa -out "$KEY_PATH/photoprism.local.issuer.key" 4096

openssl req -x509 -new -nodes -key "$KEY_PATH/photoprism.local.issuer.key" -sha256 -days 365 -out "$CERTS_PATH/photoprism.local.issuer.pem" -passin pass: -passout pass: -config "$CONF_PATH/ca.conf"

openssl x509 -outform der -in "$CERTS_PATH/photoprism.local.issuer.pem" -out "$CERTS_PATH/photoprism.local.issuer.crt"

# Generate server certificates.

echo "Generating certificate for photoprism.local..."

openssl genrsa -out "$KEY_PATH/photoprism.local.key" 4096

openssl req -new -config "$CONF_PATH/local-csr.conf" -key "$KEY_PATH/photoprism.local.key" -out "$CERTS_PATH/photoprism.local.csr"

openssl x509 -req -in "$CERTS_PATH/photoprism.local.csr" -CA "$CERTS_PATH/photoprism.local.issuer.pem" -CAkey "$KEY_PATH/photoprism.local.issuer.key" -CAcreateserial \
-out "$CERTS_PATH/photoprism.local.crt" -days 365 -sha256 -extfile "$CONF_PATH/local.conf"

openssl pkcs12 -export -in "$CERTS_PATH/photoprism.local.crt" -inkey "$KEY_PATH/photoprism.local.key" -out "$KEY_PATH/photoprism.local.pfx" -passin pass: -passout pass:

echo "Generating certificate for photoprism.me..."

openssl genrsa -out "$KEY_PATH/photoprism.me.key" 4096

openssl req -new -config "$CONF_PATH/me-csr.conf" -key "$KEY_PATH/photoprism.me.key" -out "$CERTS_PATH/photoprism.me.csr"

openssl x509 -req -in "$CERTS_PATH/photoprism.me.csr" -CA "$CERTS_PATH/photoprism.local.issuer.pem" -CAkey "$KEY_PATH/photoprism.local.issuer.key" -CAcreateserial \
-out "$CERTS_PATH/photoprism.me.crt" -days 365 -sha256 -extfile "$CONF_PATH/me.conf"

openssl pkcs12 -export -in "$CERTS_PATH/photoprism.me.crt" -inkey "$KEY_PATH/photoprism.me.key" -out "$KEY_PATH/photoprism.me.pfx" -passin pass: -passout pass:

# Change key permissions.

echo "Updating permissions of keys in '$KEY_PATH'..."

chown -R root:ssl-cert "$KEY_PATH"
chmod -R u=rwX,g=rX,o-rwx "$KEY_PATH"

# Run "update-ca-certificates".

echo "Running 'update-ca-certificates'..."
update-ca-certificates

echo "Done."