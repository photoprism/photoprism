TF_VERSION=1.15.2

# -march see https://gcc.gnu.org/onlinedocs/gcc-4.8.5/gcc/i386-and-x86-64-Options.html

all: libtensorflow
patch:
	git apply tensorflow-$(TF_VERSION).diff
download:
	wget https://github.com/tensorflow/tensorflow/archive/v$(TF_VERSION).tar.gz
	tar -xzf v$(TF_VERSION).tar.gz
	cp .tf_configure.bazelrc Makefile ../*.sh tensorflow-$(TF_VERSION)
libtensorflow:
	bazel build --jobs 2 --config=opt //tensorflow:libtensorflow.so
	./create_archive.sh jetson-nano $(TF_VERSION)
