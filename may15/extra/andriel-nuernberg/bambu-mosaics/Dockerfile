FROM golang:1.4.2

RUN mkdir -p /go/src/app/mosaics
WORKDIR /go/src/app
COPY . /go/src/app

RUN go install

VOLUME ["/go/src/app/tiles"]
EXPOSE 4000

CMD ["app"]
