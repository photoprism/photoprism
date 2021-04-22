FROM photoprism/development:20210422

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .