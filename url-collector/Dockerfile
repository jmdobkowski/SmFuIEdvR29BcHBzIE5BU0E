FROM golang:1.19-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./
RUN go build -o /server cmd/server/main.go

ENV API_KEY=DEMO_KEY PORT=8080 CONCURRENT_REQUESTS=5

EXPOSE 8080

CMD [ "/server" ]