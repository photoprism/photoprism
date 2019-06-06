FROM photoprism/development:20190606

# Install Python and TensorFlow
RUN apt-get update && apt-get install -y --no-install-recommends \
        python3 \
        python-setuptools \
        python3-dev \
        && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN curl -O https://bootstrap.pypa.io/get-pip.py && \
    python get-pip.py && \
    rm get-pip.py

RUN pip --no-cache-dir install --upgrade \
        tensorflow \
        requests \
        Pillow \
        h5py \
        ipykernel \
        jupyter \
        keras_applications \
        keras_preprocessing \
        matplotlib \
        numpy \
        pandas \
        scipy \
        sklearn \
        && \
    python -m ipykernel.kernelspec

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .

# Build PhotoPrism
RUN make all install

# Start PhotoPrism server
CMD photoprism start
