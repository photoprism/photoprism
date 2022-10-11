#!/usr/bin/env bash

SCRIPT_DIR=$(dirname "$0")

"$SCRIPT_DIR/create-ca.sh"
"$SCRIPT_DIR/create-certs.sh"