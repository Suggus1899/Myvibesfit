FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Los dos binarios salen de la misma imagen: comparten capas y evitan que el
# worker acabe desplegado con una version distinta a la del servidor.
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api && \
    CGO_ENABLED=0 go build -o /out/worker ./cmd/worker

# goose se instala aparte a proposito: es una herramienta de operacion, no una
# dependencia del modulo, y meterla en go.mod la arrastraria a cada build.
RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.25.0

FROM alpine:3.20
# tzdata no es opcional: el recordatorio de rachas de SDD-003 razona sobre la
# zona horaria del usuario, y una imagen sin base de zonas resuelve todo a UTC.
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/api /usr/local/bin/api
COPY --from=build /out/worker /usr/local/bin/worker
COPY --from=build /go/bin/goose /usr/local/bin/goose
COPY --from=build /src/db/migrations /migrations
EXPOSE 8080
ENTRYPOINT ["api"]
