FROM alpine:latest

WORKDIR /app

# Build-Arg from workflow
ARG TARGETARCH
COPY dist/fshare-linux-${TARGETARCH} /app/fshare.sh

ENTRYPOINT ["/app/fshare.sh"]