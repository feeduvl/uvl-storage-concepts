FROM ubuntu:20.04

RUN apt-get update && apt-get install -y mongodb

RUN apt-get install golang

WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

RUN go test -cover

EXPOSE 9684

CMD ["app"]