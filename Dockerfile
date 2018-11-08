FROM photoprism/development:20181108

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .

# Build PhotoPrism
RUN make all install

# Start PhotoPrism server
CMD photoprism start