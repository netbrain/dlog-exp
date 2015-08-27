FROM golang:1.5

ENV GO15VENDOREXPERIMENT 1

ADD . $GOPATH/src/github.com/netbrain/dlog-exp

RUN go install github.com/netbrain/dlog-exp/cmd/server
RUN go install github.com/netbrain/dlog-exp/cmd/webserver
