TF_VERSION=1.14.0

all: libtensorflow static archive
patch:
	git apply tensorflow-$(TF_VERSION).diff
libtensorflow:
	bazel build --jobs 2 --config=opt //tensorflow:libtensorflow.so
static:
	env JOB_COUNT=2 ARCH=core-avx-i ./build_static.sh
archive:
	mkdir -p archive/lib/
	mkdir -p archive/include/tensorflow/c/eager/
	cp bazel-bin/tensorflow/libtensorflow.so.$(TF_VERSION) archive/lib/libtensorflow.so
	cp bazel-bin/tensorflow/libtensorflow_framework.so.$(TF_VERSION) archive/lib/libtensorflow_framework.so
	cp tensorflow/c/eager/c_api.h archive/include/tensorflow/c/eager/
	cp tensorflow/c/c_api.h tensorflow/c/c_api_experimental.h tensorflow/c/LICENSE archive/include/tensorflow/c/
	(cd archive && tar -czf ../libtensorflow-cpu-linux-core-avx-i-$(TF_VERSION).tar.gz)
	du -h libtensorflow-cpu-linux-core-avx-i-$(TF_VERSION).tar.gz
