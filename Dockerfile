FROM alpine:latest

WORKDIR /app

# Build-Arg from workflow
ARG TARGETARCH
COPY dist/binaries-fshare-linux-${TARGETARCH}/fshare-linux-${TARGETARCH} /app/fshare

ENTRYPOINT ["/app/fshare"]