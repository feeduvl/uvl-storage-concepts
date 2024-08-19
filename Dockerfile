FROM golang:1.15

WORKDIR /go/src/app

COPY . .

# Initialize the module and add specific version of gorilla/mux
RUN go mod init ri-storage-twitter && \
    go get github.com/gorilla/mux@v1.8.0 && \
    go mod tidy

RUN go build -o app .

RUN ls -la /go/src/app

EXPOSE 9682

CMD ["./app"]
