FROM photoprism/development:20191214

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .
