# build
FROM golang:1.22-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /events-service ./main.go

# runtime
FROM alpine:3.20
WORKDIR /app
COPY --from=build /events-service /app/events-service
EXPOSE 8082
ENV PORT=8082
CMD ["/app/events-service"]
