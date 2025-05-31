FROM alpine:latest

WORKDIR /app

# Build-Arg from workflow
ARG TARGETARCH
COPY dist/fshare-linux-${TARGETARCH} ./fshare

ENTRYPOINT ["./fshare"]