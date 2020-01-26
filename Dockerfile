FROM photoprism/development:20200126

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
COPY . .
