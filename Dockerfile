FROM golang:1.21-alpine

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o frappuccino

EXPOSE 8080

CMD ["./frappuccino"]