FROM golang:1.14

ENV GOFLAGS=-mod=vendor

COPY . /go/src/github.com/TommyStarK/fabric-go-client

WORKDIR /go/src/github.com/TommyStarK/fabric-go-client

CMD ["/bin/bash"]
