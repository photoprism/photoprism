FROM photoprism/development:20181109

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .