FROM golang:1.4.2

ADD . /go/src/github.com/captncraig/mosaicChallenge

RUN go install github.com/captncraig/mosaicChallenge/web
RUN cp -avr /go/src/github.com/captncraig/mosaicChallenge/web/static/ /go/bin/static
RUN cp -avr /go/src/github.com/captncraig/mosaicChallenge/web/templates/ /go/bin/templates

WORKDIR /go/bin
ENTRYPOINT /go/bin/web

EXPOSE 7777