FROM golang:alpine AS build-env
RUN apk add --no-cache --update git
RUN go get -u github.com/golang/dep/...
ADD . /go/src/github.com/cv/sd
RUN cd /go/src/github.com/cv/sd && dep ensure && go build -o sd

FROM alpine
WORKDIR /app
COPY --from=build-env /go/src/github.com/cv/sd/sd /app/
ENTRYPOINT ./sd $@
