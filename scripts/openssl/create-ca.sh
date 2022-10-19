#!/usr/bin/env bash

# To add this certificate to your list of trusted issuers:
# sudo cp storage/config/certificates/photoprism.local.issuer.crt /usr/local/share/ca-certificates/photoprism.local.issuer.crt
# sudo update-ca-certificates

# shellcheck disable=SC2164
SCRIPT_PATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
CERTS_PATH="${SCRIPT_PATH}/../../storage/config/certificates"

echo "OpenSSL Scripts: ${SCRIPT_PATH}"
echo "HTTPS Cert Path: ${CERTS_PATH}"

mkdir -p "${CERTS_PATH}"

openssl genrsa -out "$CERTS_PATH/photoprism.local.issuer.key" 4096

openssl req -x509 -new -nodes -key "$CERTS_PATH/photoprism.local.issuer.key" -sha256 -days 365 -out "$CERTS_PATH/photoprism.local.issuer.pem" -passin pass: -passout pass: -config "$SCRIPT_PATH/ca.conf"

openssl x509 -outform der -in "$CERTS_PATH/photoprism.local.issuer.pem" -out "$CERTS_PATH/photoprism.local.issuer.crt"
