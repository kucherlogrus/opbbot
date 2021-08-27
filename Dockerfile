FROM golang:1.17-alpine

RUN apk add build-base

WORKDIR /app

COPY /app/go.mod /app/go.mod
COPY /app/go.sum /app/go.sum

RUN go mod download

COPY /app/lib /app/lib
COPY /app/main.go /app/main.go


RUN go build -o /opb_bot

CMD [ "/opb_bot" ]
