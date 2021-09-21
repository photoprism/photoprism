FROM photoprism/development:20210921

# Set up project directory
WORKDIR "/go/src/github.com/photoprism/photoprism"
ENV HOME="/go/src/github.com/photoprism/photoprism"
COPY . .