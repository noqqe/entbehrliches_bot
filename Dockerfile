FROM golang:1.20

# ARG APITOKEN
# ARG POSTS
# ARG GITHUB_REPO
# ARG GITHUB_OWNER
# ARG GITHUB_TOKEN
ENV APITOKEN ${APITOKEN}
ENV POSTS ${POSTS}
ENV GITHUB_REPO ${GITHUB_REPO}
ENV GITHUB_OWNER ${GITHUB_OWNER}
ENV GITHUB_TOKEN ${GITHUB_TOKEN}

WORKDIR /go/src/app
COPY . /go/src/app

# Run bamse bot
CMD [ "./bamse" ]
