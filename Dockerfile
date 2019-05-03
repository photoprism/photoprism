FROM photoprism/development:20190503

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .
