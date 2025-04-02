# -march see https://gcc.gnu.org/onlinedocs/gcc-4.8.5/gcc/i386-and-x86-64-Options.html

all: libtensorflow
download:
	wget https://github.com/tensorflow/tensorflow/archive/v$(TF_VERSION).tar.gz
	tar -xzf v$(TF_VERSION).tar.gz
	cp Makefile *.sh tensorflow-$(TF_VERSION)
libtensorflow:
	bazel build -c opt //tensorflow:libtensorflow.so --copt=-march=armv8-a
	./create_archive.sh jetson-nano $(TF_VERSION)
