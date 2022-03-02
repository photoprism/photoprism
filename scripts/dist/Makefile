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
	apt-get update
	apt-get -qq dist-upgrade
clean:
	apt-get -y autoremove
	apt-get -y autoclean
	rm -rf /var/lib/apt/lists/*
gpu:
	./install-gpu.sh
tensorflow:
	./install-tensorflow.sh auto
davfs:
	./install-davfs.sh
nano:
	apt-get update
	apt-get -qq install nano

.PHONY: update apt-upgrade clean apt-cleanup gpu tensorflow davfs nano \
    tensorflow-amd64-cpu tensorflow-amd64-avx tensorflow-amd64-avx2 \
    intel-graphics install-intel-graphics install-davfs;
