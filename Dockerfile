FROM photoprism/development:20210831

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .