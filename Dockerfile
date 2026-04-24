ARG GO_VERSION=1

FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod ./
RUN go mod download && go mod verify
COPY server/ ./server/
COPY speedDaemon/ ./speedDaemon/
COPY lineReversal/ ./lineReversal/
COPY main.go .
RUN go build -v -o /run-app .


FROM debian:bookworm

COPY --from=builder /run-app /usr/local/bin/
EXPOSE  10000
EXPOSE  10001
EXPOSE  10002
EXPOSE  10003
EXPOSE  10004
EXPOSE  10005
EXPOSE  10006
EXPOSE  10007

CMD ["run-app"]
