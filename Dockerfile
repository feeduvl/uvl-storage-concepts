FROM ubuntu:20.04

RUN apt-get update && apt-get install -y mongodb

WORKDIR /tmp/go
RUN wget https://dl.google.com/go/go1.15.4.linux-amd64.tar.gz
RUN tar -xvf go1.15.4.linux-amd64.tar.gz
RUN mv go /usr/local
RUN export GOROOT=/usr/local/go
RUN export GOPATH=$HOME/gopojects/go
RUN export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

RUN go test -cover

EXPOSE 9684

CMD ["app"]