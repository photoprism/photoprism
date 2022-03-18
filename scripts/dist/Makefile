# INSTALLS OPTIONAL PACKAGES AND DRIVERS IN DOCKER IMAGES
# Maintainer: Michael Mayer <hello@photoprism.app>

intel-graphics: gpu
install-intel-graphics: gpu
install-davfs: davfs
tensorflow-amd64-cpu: tensorflow
tensorflow-amd64-avx: tensorflow
tensorflow-amd64-avx2: tensorflow
apt-cleanup: clean
apt-upgrade: update
update:
	/usr/bin/apt-get update
	/usr/bin/apt-get -qq dist-upgrade
clean:
	/usr/bin/apt-get -y autoremove
	/usr/bin/apt-get -y autoclean
	/bin/rm -rf /var/lib/apt/lists/*
gpu:
	/scripts/install-gpu.sh
tensorflow:
	/scripts/install-tensorflow.sh auto
davfs:
	/scripts/install-davfs.sh
nano: clitools
clitools:
	/usr/bin/apt-get update
	@/usr/bin/apt-get -qq install zsh nano >/dev/null 2>&1 && echo "init: successfully installed zsh and nano" || echo "init: packages zsh and nano not available for installation"
	@/usr/bin/apt-get -qq install exa duf >/dev/null 2>&1 && echo "init: successfully installed exa and duf" || echo "init: packages exa and duf not available for installation"

.PHONY: update apt-upgrade clean apt-cleanup gpu tensorflow davfs \
    tensorflow-amd64-cpu tensorflow-amd64-avx tensorflow-amd64-avx2 \
    intel-graphics install-intel-graphics install-davfs nano clitools;
