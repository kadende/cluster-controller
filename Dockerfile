# Compile stage
FROM golang:1.10.3-alpine3.7 as builder-env

RUN apk update \
    && apk add gcc git openssh curl make libc6-compat musl-dev\
    && DEP_RELEASE_TAG=v0.4.1 curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh\
    && go get github.com/derekparker/delve/cmd/dlv

COPY .  /go/src/github.com/kadende/cluster-controller
WORKDIR /go/src/github.com/kadende/cluster-controller

# one of the plugin caveats
# https://github.com/alperkose/golangplugins#caveats
RUN go get github.com/kadende/kadende-interfaces/spi

RUN dep ensure -v \
    # The -gcflags "all=-N -l" flag helps us get a better debug experience
    && go build -gcflags "all=-N -l" -o /cluster-controller


# Final stage
FROM alpine:3.7
# Allow delve to run on Alpine based containers.
RUN apk --no-cache add ca-certificates libc6-compat
# Port 40000 belongs to Delve
EXPOSE  40000

WORKDIR /
COPY --from=builder-env /cluster-controller /
COPY --from=builder-env /go/bin/dlv /
# Run delve
CMD ["/dlv", "--listen=:40000", "--headless=true", "--api-version=2", "exec", "/cluster-controller"]