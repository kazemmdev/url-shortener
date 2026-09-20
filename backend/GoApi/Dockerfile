FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/goapi .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/goapi .

ENV HTTP_PORT=8080
EXPOSE 8080
USER nobody

ENTRYPOINT ["./goapi"]
