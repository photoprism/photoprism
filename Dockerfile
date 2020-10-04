FROM photoprism/development:20201004

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .
