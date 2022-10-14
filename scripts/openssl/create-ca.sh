#!/usr/bin/env bash

# shellcheck disable=SC2164
SCRIPT_PATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
CERTS_PATH="${SCRIPT_PATH}/../../storage/config/certs"

echo "OpenSSL Scripts: ${SCRIPT_PATH}"
echo "HTTPS Cert Path: ${CERTS_PATH}"

mkdir -p "${CERTS_PATH}"

openssl genrsa -out "$CERTS_PATH/ca.key" 4096

openssl req -x509 -new -nodes -key "$CERTS_PATH/ca.key" -sha256 -days 365 -out "$CERTS_PATH/ca.pem" -passin pass: -passout pass: -config "$SCRIPT_PATH/ca.conf"

openssl x509 -outform der -in "$CERTS_PATH/ca.pem" -out "$CERTS_PATH/ca.crt"

# To add this to the local cert list:
# sudo cp ./certs/ca.crt /usr/local/share/ca-certificates/local-ca.crt
# sudo update-ca-certificates