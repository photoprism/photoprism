FROM photoprism/development:20181218

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .