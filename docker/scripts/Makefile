# INSTALLS OPTIONAL PACKAGES AND DRIVERS IN DOCKER IMAGES
# Maintainer: Michael Mayer <hello@photoprism.app>

.PHONY: apt-upgrade apt-cleanup install-updates install-intel-graphics intel-graphics \
		tensorflow-amd64-cpu tensorflow-amd64-cpu-install \
		tensorflow-amd64-avx tensorflow-amd64-avx-install \
		tensorflow-amd64-avx2 tensorflow-amd64-avx2-install;

install-updates: apt-upgrade apt-cleanup
intel-graphics: apt-upgrade install-intel-graphics apt-cleanup
tensorflow-amd64-cpu: /tmp/libtensorflow-linux-cpu-1.15.2.tar.gz tensorflow-amd64-cpu-install
tensorflow-amd64-avx: /tmp/libtensorflow-linux-avx-1.15.2.tar.gz tensorflow-amd64-avx-install
tensorflow-amd64-avx2: /tmp/libtensorflow-linux-avx2-1.15.2.tar.gz tensorflow-amd64-avx2-install
apt-upgrade:
	apt-get update
	apt-get dist-upgrade
apt-cleanup:
	apt-get -y autoremove && apt-get -y autoclean && apt-get clean && rm -rf /var/lib/apt/lists/*
install-intel-graphics:
	apt-get install intel-opencl-icd intel-media-va-driver-non-free i965-va-driver-shaders libmfx1
/tmp/libtensorflow-linux-cpu-1.15.2.tar.gz:
	curl -fsSL "https://dl.photoprism.app/tensorflow/linux/libtensorflow-linux-cpu-1.15.2.tar.gz" > /tmp/libtensorflow-linux-cpu-1.15.2.tar.gz
tensorflow-amd64-cpu-install:
	tar --overwrite -C "/usr" -xzf /tmp/libtensorflow-linux-cpu-1.15.2.tar.gz
	ldconfig
/tmp/libtensorflow-linux-avx-1.15.2.tar.gz:
	curl -fsSL "https://dl.photoprism.app/tensorflow/linux/libtensorflow-linux-avx-1.15.2.tar.gz" > /tmp/libtensorflow-linux-avx-1.15.2.tar.gz
tensorflow-amd64-avx-install:
	tar --overwrite -C "/usr" -xzf /tmp/libtensorflow-linux-avx-1.15.2.tar.gz
	ldconfig
/tmp/libtensorflow-linux-avx2-1.15.2.tar.gz:
	curl -fsSL "https://dl.photoprism.app/tensorflow/linux/libtensorflow-linux-avx2-1.15.2.tar.gz" > /tmp/libtensorflow-linux-avx2-1.15.2.tar.gz
tensorflow-amd64-avx2-install:
	tar --overwrite -C "/usr" -xzf /tmp/libtensorflow-linux-avx2-1.15.2.tar.gz
	ldconfig
