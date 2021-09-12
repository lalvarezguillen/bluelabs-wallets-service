FROM golang:buster

ENV GO111MODULE=on

WORKDIR /api

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz \
    && mv ./migrate.linux-amd64 /usr/bin/go-migrate

COPY . .
RUN chmod +x ./entrypoint.sh
RUN go build -o api .

CMD ["./entrypoint.sh"]