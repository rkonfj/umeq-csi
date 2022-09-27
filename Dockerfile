FROM golang:1.19-alpine3.16 as builder
ADD . /app/
ENV GOPROXY https://goproxy.cn
RUN cd /app/cmd/plugin/;go build -ldflags "-s -w"

FROM alpine:3.16.2
WORKDIR /app
RUN sed -i 's/dl-cdn.alpinelinux.org/opentuna.cn/g' /etc/apk/repositories
RUN apk add e2fsprogs
COPY --from=builder /app/cmd/plugin/plugin /app/
