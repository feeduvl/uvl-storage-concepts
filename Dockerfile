
FROM golang:1.9
WORKDIR /go/src/app
COPY . .
RUN go get -t -d -v ./...
RUN go install -v ./...
RUN go test -coverprofile=coverage.out

EXPOSE 9684
CMD ["app"]