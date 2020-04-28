TF_VERSION=1.15.2

# -march see https://gcc.gnu.org/onlinedocs/gcc-4.8.5/gcc/i386-and-x86-64-Options.html

all: libtensorflow
patch:
	git apply tensorflow-$(TF_VERSION).diff
libtensorflow:
	bazel build --jobs 2 --config=opt //tensorflow:libtensorflow.so --copt=-march=armv8-a --host_copt=-march=armv8-a
	./create_archive.sh arm64 $(TF_VERSION)
