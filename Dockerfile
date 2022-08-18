FROM golang:latest AS build

WORKDIR /go/src/github.com/egimbernat/washer
COPY Makefile .

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY ./tools/washer ./tools/washer
RUN make tools/washer/washer

FROM debian:bullseye

WORKDIR /
COPY --from=build /go/src/gitlab.com/github.com/egimbernat/washer/washer /usr/local/bin/washer

RUN apt update && apt install -y wget
RUN wget https://packagecloud.io/install/repositories/wasmCloud/core/script.deb.sh
RUN bash script.deb.sh
RUN apt install -y wash jq

ENTRYPOINT ["/usr/local/bin/washer"]
