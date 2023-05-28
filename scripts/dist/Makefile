# INSTALLS OPTIONAL PACKAGES AND DRIVERS IN DOCKER IMAGES

export PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts"

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
	/scripts/install-https.sh
mariadb: mariadb-client
mariadb-client: install-mariadb-client
install-mariadb-client:
	/scripts/install-mariadb.sh mariadb-client
gpu: install-gpu
install-gpu:
	/scripts/install-gpu.sh
tensorflow: install-tensorflow
tensorflow-amd64-cpu: install-tensorflow
tensorflow-amd64-avx: install-tensorflow
tensorflow-amd64-avx2: install-tensorflow
install-tensorflow:
	/scripts/install-tensorflow.sh auto
davfs: install-davfs
install-davfs:
	/scripts/install-davfs.sh
intel: update install-intel
install-intel-graphics: intel
intel-graphics: intel
install-intel:
	@echo "Installing Intel Drivers..."
	apt-get -qq install intel-opencl-icd intel-media-va-driver-non-free i965-va-driver-shaders mesa-va-drivers libmfx-gen-dev va-driver-all vainfo libva-dev
clitools: update install-clitools
cli-tools: clitools
nano: clitools
install-clitools:
	@apt-get -qq install zsh nano >/dev/null 2>&1 && echo "init: successfully installed zsh and nano" || echo "init: packages zsh and nano not available for installation"
	@apt-get -qq install exa duf >/dev/null 2>&1 && echo "init: successfully installed exa and duf" || echo "init: packages exa and duf not available for installation"
.PHONY: update apt-upgrade clean apt-cleanup tensorflow install-tensorflow  \
	tensorflow-amd64-cpu tensorflow-amd64-avx tensorflow-amd64-avx2 \
    intel install-intel intel-graphics install-intel-graphics \
    davfs install-davfs nano clitools install-clitools \
    mariadb-client install-mariadb-client https install-https gpu install-gpu;
