FROM golang:1.25.0-alpine3.22 AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download
RUN GOBIN=/out go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.22.1

RUN adduser -D -H -u 10001 appuser
USER appuser

COPY --from=build /out/api /api
COPY --from=build /out/migrate /migrate
COPY migrations /migrations

EXPOSE 8080
ENTRYPOINT ["/api"]
