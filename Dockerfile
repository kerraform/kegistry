FROM golang:1.19-buster AS builder

ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /go/src/github.com/kerraform/kegistry

COPY go.mod go.mod
COPY go.sum go.sum
RUN GOPROXY='direct' go mod download

COPY . .
RUN go build -o /usr/bin/kegistry .

## Runtime

FROM gcr.io/distroless/base:3c213222937de49881c57c476e64138a7809dc54
COPY --from=builder /usr/bin/kegistry /usr/bin/kegistry

CMD ["/usr/bin/kegistry"]
