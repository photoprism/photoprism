FROM photoprism/development:20201121

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .
