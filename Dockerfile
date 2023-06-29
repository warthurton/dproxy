FROM golang:1.20.5-alpine as build 

WORKDIR /go/src/app
ADD . /go/src/app
RUN CGO_ENABLED=0 go build -o /dproxy -ldflags="-s -w" *.go

FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=build /dproxy /
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
CMD ["/dproxy", "/config/config.yaml"]