FROM photoprism/development:20200813

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .
