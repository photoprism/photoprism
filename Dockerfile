FROM photoprism/development:20201215

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .