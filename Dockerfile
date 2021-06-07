FROM golang:1.15

ARG GITHUB_TOKEN=$GITHUB_TOKEN

WORKDIR /app
COPY . ./
RUN go build .
RUN chmod +x entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]