# -march see https://gcc.gnu.org/onlinedocs/gcc-4.8.5/gcc/i386-and-x86-64-Options.html

all: libtensorflow libtensorflow-avx libtensorflow-avx2
download:
	wget https://github.com/tensorflow/tensorflow/archive/v$(TF_VERSION).tar.gz
	tar -xzf v$(TF_VERSION).tar.gz
	cp Makefile *.sh tensorflow-$(TF_VERSION)
libtensorflow:
	bazel build -c opt //tensorflow:libtensorflow.so
	./create_archive.sh linux-cpu $(TF_VERSION)
libtensorflow-avx:
	bazel build -c opt //tensorflow:libtensorflow.so --copt=-mavx
	./create_archive.sh linux-avx $(TF_VERSION)
libtensorflow-avx2:
	bazel build -c opt //tensorflow:libtensorflow.so --copt=-mavx2
	./create_archive.sh linux-avx2 $(TF_VERSION)
