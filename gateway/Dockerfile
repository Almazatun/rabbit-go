FROM golang:1.19.11-alpine3.18

WORKDIR /app

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o main ./cmd

EXPOSE 8080

CMD [ "./main" ]