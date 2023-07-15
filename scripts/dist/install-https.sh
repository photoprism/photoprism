#!/usr/bin/env bash

# Creates a default TLS certificate that can be used to enable HTTPS if no other certificate is available.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-https.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

# shellcheck disable=SC2164
CONF_PATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )/openssl"
CERTS_PATH="/etc/ssl/certs"
KEY_PATH="/etc/ssl/private"

# Abort if files already exist.

if [ -f "$CERTS_PATH/photoprism.issuer.crt" ] && [ -f "$KEY_PATH/photoprism.pfx" ]; then
    echo "Default HTTPS/TLS certificate already exists."
    exit 0
fi

echo "Creating a default HTTPS/TLS certificate."

mkdir -p "${CERTS_PATH}" "${KEY_PATH}"
groupadd -f -r -g 116 ssl-cert 1>&2

# Generate issuer (CA) certificate.

echo "Generating self-signed issuer (CA) certificate..."

openssl genrsa -out "$KEY_PATH/photoprism.issuer.key" 4096

openssl req -x509 -new -nodes -key "$KEY_PATH/photoprism.issuer.key" -sha256 -days 3650 -out "$CERTS_PATH/photoprism.issuer.pem" -passin pass: -passout pass: -config "$CONF_PATH/ca.conf"

openssl x509 -outform der -in "$CERTS_PATH/photoprism.issuer.pem" -out "$CERTS_PATH/photoprism.issuer.crt"

# Generate server certificates.

echo "Generating self-signed tls certificate..."

openssl genrsa -out "$KEY_PATH/photoprism.key" 4096

openssl req -new -config "$CONF_PATH/csr.conf" -key "$KEY_PATH/photoprism.key" -out "$CERTS_PATH/photoprism.csr"

openssl x509 -req -in "$CERTS_PATH/photoprism.csr" -CA "$CERTS_PATH/photoprism.issuer.pem" -CAkey "$KEY_PATH/photoprism.issuer.key" -CAcreateserial \
-out "$CERTS_PATH/photoprism.crt" -days 3650 -sha256 -extfile "$CONF_PATH/ext.conf"

openssl pkcs12 -export -in "$CERTS_PATH/photoprism.crt" -inkey "$KEY_PATH/photoprism.key" -out "$KEY_PATH/photoprism.pfx" -passin pass: -passout pass:

# Change key permissions.

echo "Updating permissions of keys in '$KEY_PATH'..."

chown -R root:ssl-cert "$KEY_PATH"
chmod -R u=rwX,g=rX,o-rwx "$KEY_PATH"

# Run "update-ca-certificates".

echo "Running 'update-ca-certificates'..."
update-ca-certificates
