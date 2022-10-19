#!/usr/bin/env bash

# shellcheck disable=SC2164
SCRIPT_PATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
CERTS_PATH="${SCRIPT_PATH}/../../storage/config/certificates"

mkdir -p "${CERTS_PATH}"

openssl genrsa -out "$CERTS_PATH/photoprism.local.key" 4096

openssl req -new -config "$SCRIPT_PATH/openssl.conf" -key "$CERTS_PATH/photoprism.local.key" -out "$CERTS_PATH/photoprism.local.csr"

openssl x509 -req -in "$CERTS_PATH/photoprism.local.csr" -CA "$CERTS_PATH/photoprism.local.issuer.pem" -CAkey "$CERTS_PATH/photoprism.local.issuer.key" -CAcreateserial \
-out "$CERTS_PATH/photoprism.local.crt" -days 365 -sha256 -extfile "$SCRIPT_PATH/local.conf"

openssl pkcs12 -export -in "$CERTS_PATH/photoprism.local.crt" -inkey "$CERTS_PATH/photoprism.local.key" -out "$CERTS_PATH/photoprism.local.pfx" -passin pass: -passout pass:
