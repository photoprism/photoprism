TF_VERSION=1.14.0

# -march see https://gcc.gnu.org/onlinedocs/gcc-4.8.5/gcc/i386-and-x86-64-Options.html

all: libtensorflow libtensorflow-avx libtensorflow-avx2
patch:
	git apply tensorflow-$(TF_VERSION).diff
libtensorflow:
	bazel build --jobs 2 --config=opt //tensorflow:libtensorflow.so
	./create_archive.sh linux-cpu $(TF_VERSION)
libtensorflow-avx:
	bazel build --jobs 2 --config=opt //tensorflow:libtensorflow.so --copt=-march=core-avx-i --host_copt=-march=core-avx-i
	./create_archive.sh linux-avx $(TF_VERSION)
libtensorflow-avx2:
	bazel build --jobs 2 --config=opt //tensorflow:libtensorflow.so --copt=-march=core-avx2 --host_copt=-march=core-avx2
	./create_archive.sh linux-avx2 $(TF_VERSION)
static-avx:
	env JOB_COUNT=2 ARCH=core-avx-i ./build_static.sh
static-avx2:
	env JOB_COUNT=2 ARCH=core-avx2 ./build_static.sh
