FROM golang:1.18-bullseye as builder

WORKDIR /build

COPY . .
RUN go build -v -o ./tickers ./app/



FROM debian:bullseye

RUN apt-get update \
    && apt-get install -y ca-certificates \
    \
    && groupadd --system --gid 999 app \
    && useradd --system --uid 999 --gid app app
USER app

WORKDIR /srv

COPY --chown=app:app entrypoint.sh wait-for-it.sh ./
COPY --chown=app:app --from=builder /build/tickers ./tickers

ENTRYPOINT ["/srv/entrypoint.sh"]

CMD ["/srv/tickers"]
