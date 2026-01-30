FROM golang:latest as build


WORKDIR /go/src/tun/

COPY . . 

RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN  CGO_ENABLED=0 go build -o /go/bin/tun/ -ldflags="-s -w -buildid="


#FROM --platform=linux/amd64 gcr.io/distroless/static:nonroot
FROM alpine:latest

COPY --from=build /go/src/tun/example/config.json /etc/tun/config.json
COPY --from=build /go/bin/tun /usr/local/bin/


ENTRYPOINT [ "/usr/local/bin/tun" ]
CMD [ "-c", "/etc/tun/config.json" ]
