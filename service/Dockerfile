FROM golang:1.18.0-alpine3.14

WORKDIR /app

COPY go.mod .

COPY go.sum .

COPY . .

RUN go build -o main .

WORKDIR /dist

RUN cp /app/main .

EXPOSE 3000

CMD ["/dist/main"]