#!/usr/bin/env bash

bash <(curl -s https://codecov.io/bash) -t "$CODECOV_TOKEN"
