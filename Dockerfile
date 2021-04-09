FROM golang:1.16

ARG APITOKEN
ENV APITOKEN ${APITOKEN}

ARG POSTS
ENV POSTS ${POSTS}

WORKDIR /go/src/app
COPY . /go/src/app

RUN go get -v ./...
RUN go build

# Run nichtparasoup
CMD [ "./entbehrliches_bot" ]
