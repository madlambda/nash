FROM golang:1.7

ADD . /go/src/github.com/NeowayLabs/nash

RUN cd /go/src/github.com/NeowayLabs/nash/cmd/nash && make

CMD ["nash"]
