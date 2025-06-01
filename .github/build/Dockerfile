# debian for cgo
FROM debian:bookworm-slim
# updates
RUN apt-get update && apt-get upgrade -y && apt-get clean

WORKDIR /app

ARG TARGETARCH
COPY dist/binaries-fshare-linux-${TARGETARCH}/fshare-linux-${TARGETARCH} /app/fshare

ENTRYPOINT ["/app/fshare"]