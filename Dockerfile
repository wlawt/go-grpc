# https://codefresh.io/docs/docs/learn-by-example/golang/golang-hello-world/
# https://hub.docker.com/_/golang

# FROM golang:1.15.7-buster
FROM golang:1.14

# Copy contents
WORKDIR $GOPATH/src/github.com/wlawt/goprojects/blog
COPY . .

# Install dependencies
RUN go get -d -v ./...
RUN go install -v ./...
RUN go get -u github.com/wlawt/goprojects/blog
RUN go build -o ./blog .

EXPOSE 50051

CMD ["blog"]
