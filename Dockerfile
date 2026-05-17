FROM golang:1.22-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/zetta ./cmd

FROM alpine:3.20

WORKDIR /app

COPY --from=build /bin/zetta /app/zetta
COPY db/migrations /app/db/migrations

EXPOSE 8080

ENTRYPOINT ["/app/zetta"]
