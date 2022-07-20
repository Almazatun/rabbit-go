FROM golang:1.18.0-alpine3.14

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd

WORKDIR /dist

RUN cp /app/main .

CMD ["/dist/main"]