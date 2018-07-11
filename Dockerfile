FROM golang:1.10.3-alpine as builder

RUN apk update \
    && apk add git openssh curl make\
    && DEP_RELEASE_TAG=v0.4.1 curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

COPY .  /go/src/github.com/kadende/cluster-controller
WORKDIR /go/src/github.com/kadende/cluster-controller

RUN dep ensure -v && go build -o dist/cluster-controller

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/kadende/cluster-controller/dist/cluster-controller /usr/local/bin

ENTRYPOINT ["cluster-controller"]