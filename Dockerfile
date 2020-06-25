FROM photoprism/development:20200625

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .
