# -march see https://gcc.gnu.org/onlinedocs/gcc-4.8.5/gcc/i386-and-x86-64-Options.html

all: libtensorflow
download:
	wget https://github.com/tensorflow/tensorflow/archive/v$(TF_VERSION).tar.gz
	tar -xzf v$(TF_VERSION).tar.gz
	cp Makefile *.sh tensorflow-$(TF_VERSION)
libtensorflow:
	bazel build -c opt --config cross_compile_linux_arm64 //tensorflow:libtensorflow.so
	./create_archive.sh arm64 $(TF_VERSION)
