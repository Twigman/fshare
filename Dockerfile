FROM alpine:latest

WORKDIR /app

COPY dist/fshare-linux-amd64 .

ENTRYPOINT ["./fshare-linux-amd64"]