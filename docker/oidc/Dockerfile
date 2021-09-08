FROM golang:1.17-alpine

# Move to working directory /app
WORKDIR /app

# Copy files and download dependency using go mod
COPY . .
RUN go mod download

# Build the application
RUN go build -o test-op .

# Allow HTTP scheme
ENV CAOS_OIDC_DEV=true

# Expose HTTP port
EXPOSE 9998

CMD ["/app/test-op"]