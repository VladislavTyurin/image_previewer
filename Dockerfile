from golang:1.21rc3-alpine3.18 as build

ENV BIN_FILE /opt/previewer/previewer-app
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} imagepreviewer

# На выходе тонкий образ
FROM alpine:3.18

LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="Image previewer"
LABEL MAINTAINERS="https://github.com/VladislavTyurin"

ENV BIN_FILE "/opt/previewer/previewer-app"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE /etc/previewer/config.yaml
COPY config.yaml ${CONFIG_FILE}

CMD ${BIN_FILE} -c ${CONFIG_FILE}
