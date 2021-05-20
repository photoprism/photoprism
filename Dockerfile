FROM photoprism/development:20210520

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .