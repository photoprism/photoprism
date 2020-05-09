FROM photoprism/development:20200509

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .
