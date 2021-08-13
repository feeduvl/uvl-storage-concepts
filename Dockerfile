FROM mongo:latest

FROM golang:1.15
COPY --from=golang:1.15-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /go/src/app
COPY . .
RUN go get -t -d -v ./...
RUN go install -v ./...

EXPOSE 9684

RUN go test -cover

EXPOSE 9684

CMD ["app"]