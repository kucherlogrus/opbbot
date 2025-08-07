FROM golang:1.24-alpine

RUN apk add build-base

WORKDIR /app

COPY /app/go.mod /app/go.mod
COPY /app/go.sum /app/go.sum

RUN go mod download

COPY /app/lib /app/lib
COPY /app/main.go /app/main.go

EXPOSE 8080

ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"

RUN go build -o /opb_bot

CMD [ "/opb_bot" ]
