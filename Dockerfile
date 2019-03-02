FROM golang:1.12

ADD . /go/src/github.com/NeowayLabs/nash

ENV NASHPATH /nashpath
ENV NASHROOT /nashroot

RUN cd /go/src/github.com/NeowayLabs/nash && \
    make install

CMD ["/nashroot/bin/nash"]
