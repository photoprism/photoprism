variable "TAG" {
    default = "latest"
}
variable "DOCKER_REPO" {
    default = "photoprism/photoprism"
}

group "default" {
    targets = ["arm", "amd64"]
}

group "arm" {
    targets = ["armv7", "arm64"]
}

target "armv7" {
    dockerfile = "docker/photoprism/Dockerfile"
    tags = ["${DOCKER_REPO}:${TAG}-armv7"]
    platforms = ["linux/arm/v7"]
    # No need, as this is done by the "--push" flag
    # output = ["type=registry"]
}

target "arm64" {
    dockerfile = "docker/photoprism/Dockerfile"
    tags = ["${DOCKER_REPO}:${TAG}-arm64"]
    platforms = ["linux/arm64"]
    # No need, as this is done by the "--push" flag
    # output = ["type=registry"]
}

target "amd64" {
    dockerfile = "docker/photoprism/Dockerfile"
    tags = ["${DOCKER_REPO}:${TAG}-amd64"]
    platforms = ["linux/amd64"]
    # No need, as this is done by the "--push" flag
    # output = ["type=registry"]
}
