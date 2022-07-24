FROM golang:1.18-bullseye as builder

WORKDIR /build

COPY . .

# hadolint ignore=DL3003
RUN cd app/ && go build -v -o ./tickers



FROM debian:bullseye-slim

# hadolint ignore=DL3008,DL3009
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
    \
    && groupadd --system --gid 999 app \
    && useradd --system --uid 999 --gid app app
USER app

WORKDIR /srv

COPY --chown=app:app --from=builder /build/app/tickers ./

CMD ["/srv/tickers"]
