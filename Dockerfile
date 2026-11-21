# Build stage: compiles a static Go binary.
# The "1.27-alpine3.24" tag picks up Go patch releases automatically. Trivy also
# scans the Go standard library inside the binary, so this matters.
FROM golang:1.27-alpine3.24 AS build
WORKDIR /src
COPY go.mod *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/rickroll .

# Runtime stage: slim Alpine image
FROM alpine:3.24
# Install security updates that are newer than the base image
RUN apk upgrade --no-cache
COPY --from=build /out/rickroll /usr/local/bin/rickroll
USER nobody
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/rickroll"]
