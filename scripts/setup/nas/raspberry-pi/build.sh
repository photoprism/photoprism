#!/usr/bin/env bash

# Builds a PhotoPrismPi SD card image for use with Raspberry Pi 4 and 5.

# Stop the script if an error occurs.
set -e

echo "Building PhotoPrismPi SD card image..."

# Build directory:
DESTDIR=$(realpath "${1:-./setup/nas/raspberry-pi}")

# Ubuntu Server version and download URL:
UBUNTU_VERSION="${2:-24.04.3}"
UBUNTU_URL="https://cdimage.ubuntu.com/releases/${UBUNTU_VERSION}/release/ubuntu-${UBUNTU_VERSION}-preinstalled-server-arm64+raspi.img.xz"

# SD card image file name and path:
IMAGE_NAME="photoprismpi-ubuntu-${UBUNTU_VERSION}.img"
IMAGE_PATH="${DESTDIR}/${IMAGE_NAME}"

# Cloud init config path:
CONFIG_PATH="${DESTDIR}/cloud-init"

# Boot partition mount path:
MOUNT_DEV="/dev/nbd0"
MOUNT_PATH="${DESTDIR}/boot"

# Show image and build details.
echo "--------------------------------------------------------------------------------"
echo "VERSION: Ubuntu Server ${UBUNTU_VERSION} for Raspberry Pi"
echo "CDIMAGE: ${UBUNTU_URL}"
echo "DESTDIR: ${DESTDIR}"
echo "SDIMAGE: ${IMAGE_PATH}.xz"
echo "--------------------------------------------------------------------------------"

# Install build dependencies.
sudo apt update
sudo apt install -y qemu-utils xz-utils cloud-init

# Remove existing Ubuntu Server image, if any.
rm -f "${IMAGE_PATH}" "${IMAGE_PATH}.xz"

# Download latest Ubuntu Server image.
echo "Downloading Ubuntu Server image..."
curl -o "${IMAGE_PATH}.xz" -fsSL "${UBUNTU_URL}"
echo "Done."

# Unpack Ubuntu Server image.
echo "Unpacking ${IMAGE_NAME}.xz..."
(cd "${DESTDIR}" && unxz "${IMAGE_NAME}.xz")
echo "Done."

# Mount the boot partition to customize it.
echo "Mounting boot partition to ${MOUNT_PATH}..."
mkdir -p "${MOUNT_PATH}"
sudo umount -q "${MOUNT_PATH}" || true

if [[ -e "${MOUNT_DEV}p1" ]]; then
  sudo qemu-nbd --disconnect "${MOUNT_DEV}" || true
fi

sleep 1
sudo modprobe nbd max_part=8
sudo qemu-nbd --connect="${MOUNT_DEV}" --format=raw "${IMAGE_PATH}"
sleep 3
sudo mount "${MOUNT_DEV}p1" "${MOUNT_PATH}"
echo "Done."

# Copy cloud-init files to the boot partition.
echo "Copying files to boot partition..."
sudo cp "${CONFIG_PATH}/meta-data" "${CONFIG_PATH}/network-config" "${CONFIG_PATH}/user-data" "${MOUNT_PATH}"
echo "Done."

# Unmount boot partition.
echo "Unmounting boot partition..."
sudo umount "${MOUNT_PATH}"
sleep 1
if [[ -e "${MOUNT_DEV}p1" ]]; then
  sudo qemu-nbd --disconnect "${MOUNT_DEV}"
fi
sleep 1
rmdir "${MOUNT_PATH}"
echo "Done."

# Create the final SD card image.
echo "Creating ${IMAGE_PATH}.xz..."
xz -T0 -z -q -9 "${IMAGE_PATH}"
echo "Done."