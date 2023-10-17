FROM harbor-registry.inner.youdao.com/infraop/golang:1.18beta2-alpine3.15-make
RUN apk --no-cache add ca-certificates
RUN apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /go/src/httpproxy
COPY . .
RUN ls .
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn
RUN go mod vendor
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -mod=vendor -buildvcs=false -a -ldflags="-s -w"  -o httpproxy .

WORKDIR /go/src/httpproxy
CMD ["-c","proxy.json"]
ENTRYPOINT ["./httpproxy"]
