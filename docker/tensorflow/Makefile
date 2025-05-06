all: libtensorflow libtensorflow-ssse3 libtensorflow-avx libtensorflow-avx2 libtensorflow-vnni libtensorflow-avx512
# Downloads the TensorFlow source code from GitHub and extracts it to a subdirectory:
download:
	wget https://github.com/tensorflow/tensorflow/archive/v$(TF_VERSION).tar.gz
	tar -xzf v$(TF_VERSION).tar.gz
	cp Makefile *.sh tensorflow-$(TF_VERSION)
# Clang command line argument reference:
# https://clang.llvm.org/docs/ClangCommandLineReference.html#x86
libtensorflow:
	bazel build -c opt //tensorflow:libtensorflow.so --copt=-O2 --copt=-mno-avx --copt=-mno-avx2 --copt=-mno-fma --copt=-march=x86-64
	./create_archive.sh amd64 $(TF_VERSION)
libtensorflow-ssse3:
	bazel build -c opt //tensorflow:libtensorflow.so --copt=-O2 --copt=-mno-avx --copt=-mno-avx2 --copt=-mno-fma --copt=-march=core2
	./create_archive.sh amd64-ssse3 $(TF_VERSION)
libtensorflow-avx:
	bazel build -c opt //tensorflow:libtensorflow.so --copt=-O2 --copt=-mavx
	./create_archive.sh amd64-avx $(TF_VERSION)
libtensorflow-avx2:
	bazel build -c opt //tensorflow:libtensorflow.so --copt=-O2 --copt=-mavx2 --copt=-mfma --copt=-mfpmath=both --copt=-msse4.2
	./create_archive.sh amd64-avx2 $(TF_VERSION)
libtensorflow-vnni:
	bazel build -c opt //tensorflow:libtensorflow.so --copt=-O2 --copt=-mavx2 --copt=-mavxvnni --copt=-mfma --copt=-mfpmath=both --copt=-msse4.2
	./create_archive.sh amd64-vnni $(TF_VERSION)
libtensorflow-avx512:
	bazel build -c opt //tensorflow:libtensorflow.so --copt=-O2 --copt=-mavx2 --copt=-mavx512f --copt=-mavx512vnni --copt=-mfma --copt=-mfpmath=both --copt=-msse4.2
	./create_archive.sh amd64-avx512 $(TF_VERSION)
