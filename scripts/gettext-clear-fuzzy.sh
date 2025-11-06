#!/usr/bin/env bash
set -euo pipefail

shopt -s globstar nullglob

remove_fuzzy_flag() {
	local file="$1"
	# Skip files that do not contain the fuzzy marker.
	if ! grep -q '^#,\ fuzzy$' "$file"; then
		return
	fi

	local tmp
	tmp="$(mktemp)"
	# Copy every line except the fuzzy marker.
	awk '$0 != "#, fuzzy"' "$file" >"$tmp"
	mv "$tmp" "$file"
}

echo "Removing fuzzy attribute from backend translations..."
for file in ./assets/locales/**/*.po; do
	remove_fuzzy_flag "$file"
done

echo "Removing fuzzy attribute from frontend translations..."
for file in ./frontend/src/locales/*.po; do
	remove_fuzzy_flag "$file"
done

echo "Done."
