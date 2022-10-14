#!/usr/bin/env bash

# shellcheck disable=SC2164
SCRIPT_PATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
CERTS_PATH="${SCRIPT_PATH}/../../storage/config/certs"

echo "OpenSSL Scripts: ${SCRIPT_PATH}"
echo "HTTPS Cert Path: ${CERTS_PATH}"

mkdir -p "${CERTS_PATH}"

openssl genrsa -out "$CERTS_PATH/local.key" 4096

openssl req -new -config "$SCRIPT_PATH/openssl.conf" -key "$CERTS_PATH/local.key" -out "$CERTS_PATH/local.csr"

openssl x509 -req -in "$CERTS_PATH/local.csr" -CA "$CERTS_PATH/ca.pem" -CAkey "$CERTS_PATH/ca.key" -CAcreateserial \
-out "$CERTS_PATH/local.crt" -days 365 -sha256 -extfile "$SCRIPT_PATH/local.conf"

openssl pkcs12 -export -in "$CERTS_PATH/local.crt" -inkey "$CERTS_PATH/local.key" -out "$CERTS_PATH/local.pfx" -passin pass: -passout pass: