TF_VERSION=1.14.0

all: libtensorflow static archive
patch:
	git apply tensorflow-$(TF_VERSION).diff
libtensorflow:
	bazel build --jobs 2 --config=opt //tensorflow:libtensorflow.so
static:
	env JOB_COUNT=2 ARCH=core-avx-i ./build_static.sh
archive:
	rm -rf tmp
	mkdir -p tmp/lib/
	mkdir -p tmp/include/tensorflow/c/eager/
	cp bazel-bin/tensorflow/libtensorflow.so.$(TF_VERSION) tmp/lib/libtensorflow.so
	cp bazel-bin/tensorflow/libtensorflow_framework.so.$(TF_VERSION) tmp/lib/libtensorflow_framework.so
	cp tensorflow/c/eager/c_api.h tmp/include/tensorflow/c/eager/
	cp tensorflow/c/c_api.h tensorflow/c/c_api_experimental.h LICENSE tmp/include/tensorflow/c/
	(cd tmp && tar -czf ../libtensorflow-linux-avx-$(TF_VERSION).tar.gz)
	du -h libtensorflow-linux-avx-$(TF_VERSION).tar.gz
