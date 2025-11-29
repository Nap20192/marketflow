FROM golang:1.24.6

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o main ./cmd

CMD [ "./main" ]
