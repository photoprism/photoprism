#!/usr/bin/env bash

echo "Removing fuzzy attribute from backend translations…"
for file in ./assets/locales/**/*.po; do msgattrib --clear-fuzzy -o "${file}" "${file}"; done

echo "Removing fuzzy attribute from frontend translations…"
for file in ./frontend/src/locales/*.po; do msgattrib --clear-fuzzy -o "${file}" "${file}"; done

echo "Done."
