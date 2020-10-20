FROM photoprism/development:20201020

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .
