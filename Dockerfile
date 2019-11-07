FROM photoprism/development:20191105

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .
