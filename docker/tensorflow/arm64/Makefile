all: libtensorflow
# Downloads the TensorFlow source code from GitHub and extracts it to a subdirectory:
download:
	wget https://github.com/tensorflow/tensorflow/archive/v$(TF_VERSION).tar.gz
	tar -xzf v$(TF_VERSION).tar.gz
	cp Makefile *.sh tensorflow-$(TF_VERSION)
# Clang command line argument reference:
# https://clang.llvm.org/docs/ClangCommandLineReference.html#aarch64
libtensorflow:
	bazel build -c opt //tensorflow:libtensorflow.so --copt=-O2 --copt=-march=armv8-a
	./create_archive.sh arm64 $(TF_VERSION)
