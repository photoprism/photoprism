# Optional packages and drivers for PhotoPrism
# Maintainer: Michael Mayer <hello@photoprism.org>
.PHONY: tensorflow-amd64-cpu tensorflow-amd64-cpu-install \
		tensorflow-amd64-avx tensorflow-amd64-avx-install \
		tensorflow-amd64-avx2 tensorflow-amd64-avx2-install \
		intel-graphics;
tensorflow-amd64-cpu: /tmp/libtensorflow-linux-cpu-1.15.2.tar.gz tensorflow-amd64-cpu-install
/tmp/libtensorflow-linux-cpu-1.15.2.tar.gz:
	curl -fsSL "https://dl.photoprism.org/tensorflow/linux/libtensorflow-linux-cpu-1.15.2.tar.gz" > /tmp/libtensorflow-linux-cpu-1.15.2.tar.gz
tensorflow-amd64-cpu-install:
	tar --overwrite -C "/usr" -xzf /tmp/libtensorflow-linux-cpu-1.15.2.tar.gz
	ldconfig
tensorflow-amd64-avx: /tmp/libtensorflow-linux-avx-1.15.2.tar.gz tensorflow-amd64-avx-install
/tmp/libtensorflow-linux-avx-1.15.2.tar.gz:
	curl -fsSL "https://dl.photoprism.org/tensorflow/linux/libtensorflow-linux-avx-1.15.2.tar.gz" > /tmp/libtensorflow-linux-avx-1.15.2.tar.gz
tensorflow-amd64-avx-install:
	tar --overwrite -C "/usr" -xzf /tmp/libtensorflow-linux-avx-1.15.2.tar.gz
	ldconfig
tensorflow-amd64-avx2: /tmp/libtensorflow-linux-avx2-1.15.2.tar.gz tensorflow-amd64-avx2-install
/tmp/libtensorflow-linux-avx2-1.15.2.tar.gz:
	curl -fsSL "https://dl.photoprism.org/tensorflow/linux/libtensorflow-linux-avx2-1.15.2.tar.gz" > /tmp/libtensorflow-linux-avx2-1.15.2.tar.gz
tensorflow-amd64-avx2-install:
	tar --overwrite -C "/usr" -xzf /tmp/libtensorflow-linux-avx2-1.15.2.tar.gz
	ldconfig
intel-graphics:
	apt-get update
	apt-get dist-upgrade
	apt-get install intel-opencl-icd intel-media-va-driver-non-free i965-va-driver-shaders libmfx1
	apt-get -y autoremove && apt-get -y autoclean && apt-get clean && rm -rf /var/lib/apt/lists/*