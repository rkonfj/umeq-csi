# Copyright 2022 rkonfj@fnla.io
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.19-alpine3.16 as builder
ADD . /app/
ENV GOPROXY https://goproxy.cn
ARG VERSION
RUN echo "build version: $VERSION";cd /app/cmd/plugin/;go build -ldflags "-s -w -X 'main.Version=$VERSION'"

FROM alpine:3.16.2
WORKDIR /app
RUN sed -i 's/dl-cdn.alpinelinux.org/opentuna.cn/g' /etc/apk/repositories
RUN apk add e2fsprogs
COPY --from=builder /app/cmd/plugin/plugin /app/
