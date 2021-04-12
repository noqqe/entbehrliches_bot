FROM golang:1.16

ENV APITOKEN ${APITOKEN}
ENV POSTS ${POSTS}
ENV GITHUB_REPO ${GITHUB_REPO}
ENV GITHUB_OWNER ${GITHUB_OWNER}
ENV GITHUB_TOKEN ${GITHUB_TOKEN}

WORKDIR /go/src/app
COPY . /go/src/app

RUN go get -v ./...
RUN go build

# Run nichtparasoup
CMD [ "./entbehrliches_bot" ]
