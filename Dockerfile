FROM golang:1.12

ADD . /go/src/github.com/madlambda/nash

ENV NASHPATH /nashpath
ENV NASHROOT /nashroot

RUN cd /go/src/github.com/madlambda/nash && \
    make install

CMD ["/nashroot/bin/nash"]
