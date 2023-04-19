FROM golang:1.20-alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates

WORKDIR /go/src/app
COPY . /go/src/app/

# build
RUN go get -v ./...
RUN GOOS=linux go build -ldflags "-X main.Version=`git describe --tags`"  -v bamse.go

# copy
FROM alpine:3.17

ENV APITOKEN ${APITOKEN}
ENV POSTS ${POSTS}
ENV GITHUB_REPO ${GITHUB_REPO}
ENV GITHUB_OWNER ${GITHUB_OWNER}
ENV GITHUB_TOKEN ${GITHUB_TOKEN}

RUN apk update && apk add --no-cache ca-certificates

WORKDIR /go/src/app
COPY --from=builder /go/src/app/bamse /go/src/app/bamse

# run
CMD [ "/go/src/app/bamse" ]
