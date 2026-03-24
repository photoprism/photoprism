#!/usr/bin/env bash

PDFJS_VERSION="4.10.38"
PDFJS_NAME="PDF.js"
PDFJS_URL="https://github.com/mozilla/pdf.js/releases/download/v${PDFJS_VERSION}/pdfjs-${PDFJS_VERSION}-dist.zip"
PDFJS_ZIP="/tmp/photoprism/pdfjs-${PDFJS_VERSION}-dist.zip"
PDFJS_DEST="assets/static/pdfjs"
PDFJS_VERSION_FILE="${PDFJS_DEST}/version.txt"

echo "Installing ${PDFJS_NAME} ${PDFJS_VERSION} viewer..."

# Already up to date?
if [[ -f "${PDFJS_VERSION_FILE}" ]] && grep -q "${PDFJS_VERSION}" "${PDFJS_VERSION_FILE}"; then
    echo "${PDFJS_NAME} ${PDFJS_VERSION} is already installed."
    exit 0
fi

# Create temp directory.
mkdir -p /tmp/photoprism

# Download release zip.
echo "Downloading ${PDFJS_NAME} ${PDFJS_VERSION} from ${PDFJS_URL}..."
wget --inet4-only -c "${PDFJS_URL}" -O "${PDFJS_ZIP}"

# Remove previous installation and extract.
rm -rf "${PDFJS_DEST}"
mkdir -p "${PDFJS_DEST}"
unzip -q "${PDFJS_ZIP}" -d "${PDFJS_DEST}"

# Write version marker.
echo "${PDFJS_VERSION}" > "${PDFJS_VERSION_FILE}"

echo "${PDFJS_NAME} ${PDFJS_VERSION} installed at ${PDFJS_DEST}."
