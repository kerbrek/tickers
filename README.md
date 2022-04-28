# Tickers

Веб-сервер, который по обращении к нему отдает данные в указанном формате.
Данные берутся из внешнего источника и сохраняются в БД раз в 30 секунд, а ответ
на запрос формируется на основе данных в БД.

Запускается командой `docker-compose up --build` и доступен по адресу
<http://127.0.0.1:8080/tickers>.

## Структура данных

Источник: <https://api.blockchain.com/v3/exchange/tickers>

На входе следущая структура:

```js
[{symbol: [string], price_24h: [float64], volume_24h: [float64], last_trade_price: [float64]}...]
```

На выходе список имеет следующий вид:

```js
{<symbol>: {price: <price_24h>, volume: <volume_24h>, last_trade: <last_trade_price>}...}
```

## Пример

Входные данные:

```json
[
  {
    "symbol":"XLM-EUR",
    "price_24h":0.25685,
    "volume_24h":49644.7076291,
    "last_trade_price":0.24
  }
]
```

Выходные данные:

```json
{
  "XLM-EUR": {
    "price": 0.25685,
    "volume": 49644.7076291,
    "last_trade":0.24
  }
}
```

## Prerequisites

- make
- docker
- docker-compose
- [dotenv](https://www.npmjs.com/package/dotenv-cli)

## Commands

- Start _Docker Compose_ services

  `make up`

- Start application

  `make start`

- Run tests

  `make test`

- List all available _Make_ commands

  `make help`
