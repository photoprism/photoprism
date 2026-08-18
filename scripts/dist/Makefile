# INSTALLS OPTIONAL PACKAGES AND DRIVERS IN DOCKER IMAGES

export PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts"

# Overview of the most commonly used targets, shown by "make help" and when no target is specified.
# Keep this in sync when renaming or adding an important target; "make list" shows all of them.
define HELP_TEXT
PhotoPrism Docker Image Makefile

Installs optional packages and drivers in PhotoPrism Docker images. Run the
targets as root with "make -C /scripts <target>", or add their names to
PHOTOPRISM_INIT so that they are installed when the container starts.

System:
  update           Update the package lists and installed packages
  clean            Remove the cached package files and package lists
  clitools         Install the zsh, nano, exa, and duf command-line tools

Security:
  https            Create a default HTTPS/TLS certificate

Graphics & Video:
  gpu              Detect and install the GPU drivers for this system
  intel            Install the Intel GPU drivers and VA-API tools
  amd              Install the AMD VA-API GPU drivers
  vulkan           Install the Mesa Vulkan drivers and tools
  ffmpeg           Install the latest FFmpeg build in "/opt"
  darktable        Install darktable for RAW image conversion
  Related: ffmpeg-release, ffmpeg-master

Machine Learning:
  tensorflow       Install the TensorFlow library for CPUs
  tensorflow-gpu   Install the TensorFlow library with GPU support
  onnxruntime      Install the ONNX Runtime library

Other:
  mariadb-client   Install the MariaDB command-line client
  proxysql-admin   Install the ProxySQL admin and scheduler tools
  davfs            Install the davfs2 WebDAV filesystem driver
  yt-dlp           Update yt-dlp to the latest release
  Related: yt-dlp-nightly, yt-dlp-master

Run "make -C /scripts list" to show all targets, including those not listed here.
endef
export HELP_TEXT

all: help
help:
	@printf '%s\n' "$$HELP_TEXT"
list:
	@awk '/^define /{skip=1} /^endef/{skip=0;next} skip{next} /^[[:alnum:]]+[^[:space:]]+:/ {printf "%s",substr($$1,1,length($$1)-1); if (match($$0,/#/)) {desc=substr($$0,RSTART+1); sub(/^[[:space:]]+/,"",desc); printf " - %s\n",desc} else printf "\n" }' "$(firstword $(MAKEFILE_LIST))"
apt-upgrade: update
update:
	apt-get update
	apt-get -qq dist-upgrade
apt-cleanup: clean
clean:
	apt-get -y autoremove
	apt-get -y autoclean
	rm -rf /var/lib/apt/lists/*
https: install-https
install-https:
	@/scripts/install-https.sh
ffmpeg: install-ffmpeg
install-ffmpeg: ffmpeg-latest
ffmpeg-release: install-ffmpeg-release
install-ffmpeg-release:
	@/scripts/install-ffmpeg.sh release
ffmpeg-latest: install-ffmpeg-latest
install-ffmpeg-latest:
	@/scripts/install-ffmpeg.sh latest /opt/ffmpeg-latest
ffmpeg-master: install-ffmpeg-master
install-ffmpeg-master:
	@/scripts/install-ffmpeg.sh master /opt/ffmpeg-master
mariadb: mariadb-client
mariadb-client: install-mariadb-client
install-mariadb-client:
	/scripts/install-mariadb.sh mariadb-client
proxysql: proxysql-admin
proxysql-admin: install-proxysql-admin
install-proxysql-admin:
	/scripts/install-proxysql-admin.sh
gpu: install-gpu
install-gpu:
	/scripts/install-gpu.sh
tensorflow: install-tensorflow
tensorflow-amd64-cpu: install-tensorflow
tensorflow-amd64-avx: install-tensorflow
tensorflow-amd64-avx2: install-tensorflow
install-tensorflow:
	/scripts/install-tensorflow.sh auto
onnxruntime: install-onnx
install-onnx:
	/scripts/install-onnx.sh
tensorflow-gpu: install-tensorflow-gpu
install-tensorflow-gpu:
	/scripts/install-tensorflow.sh gpu
codex:
	/scripts/install-codex.sh
davfs: install-davfs
install-davfs:
	/scripts/install-davfs.sh
intel: update install-intel
install-intel-graphics: intel
intel-graphics: intel
install-intel:
	@echo "Installing Intel GPU Drivers..."
	apt-get -qq install intel-opencl-icd intel-media-va-driver-non-free i965-va-driver-shaders mesa-va-drivers libmfx-gen1.2 va-driver-all vainfo libva2 libvpl2
amd: update install-amd
install-amd:
	@echo "Installing AMD VA-API GPU Drivers..."
	apt-get -qq install mesa-va-drivers vainfo libva2
vulkan: update install-vulkan
vulkan-drivers: vulkan
install-vulkan:
	@echo "Installing Mesa Vulkan Drivers and Tools..."
	apt-get -qq install mesa-vulkan-drivers vulkan-tools libvulkan1
clitools: update install-clitools
cli-tools: clitools
nano: clitools
install-clitools:
	@apt-get -qq install zsh nano >/dev/null 2>&1 && echo "init: successfully installed zsh and nano" || echo "init: packages zsh and nano not available for installation"
	@apt-get -qq install exa duf >/dev/null 2>&1 && echo "init: successfully installed exa and duf" || echo "init: packages exa and duf not available for installation"
darktable:
	echo 'deb http://download.opensuse.org/repositories/graphics:/darktable/xUbuntu_25.10/ /' | sudo tee /etc/apt/sources.list.d/graphics:darktable.list
	curl -fsSL https://download.opensuse.org/repositories/graphics:darktable/xUbuntu_25.10/Release.key | gpg --dearmor | sudo tee /etc/apt/trusted.gpg.d/graphics_darktable.gpg > /dev/null
	sudo apt-get update
	sudo apt-get install darktable
	sudo apt-get autoremove
yt-dlp: update-yt-dlp
update-yt-dlp:
	sudo yt-dlp -U
yt-dlp-nightly: install-yt-dlp-nightly
install-yt-dlp-nightly:
	@/scripts/install-yt-dlp.sh --nightly
yt-dlp-master: install-yt-dlp-master
install-yt-dlp-master:
	@/scripts/install-yt-dlp.sh --master

# Declare all targets as "PHONY", see https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html.
MAKEFLAGS += --always-make
