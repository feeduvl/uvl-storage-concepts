FROM golang:1.15-alpine

RUN apt-get update && apt-get install -y mongodb

WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

RUN go test -cover

EXPOSE 9684

CMD ["app"]