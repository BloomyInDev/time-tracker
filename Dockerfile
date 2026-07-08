FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go tool templ generate
RUN CGO_ENABLED=0 go build -o /out/time-tracker ./cmd/server
RUN CGO_ENABLED=0 go build -o /out/cli ./cmd/cli

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/time-tracker /out/cli /app/

EXPOSE 8080
ENV DB_PATH=/data/time-tracker.db
VOLUME ["/data"]

ENTRYPOINT ["/app/time-tracker"]
