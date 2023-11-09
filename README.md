# Microfibre

A bare bones Go-based API (using Gin + SQLite) for creating, updating, and
reading status updates.

Think of it as a poor-mans Twitter. No social validation, no fancy features,
just a stream of updates!

## API:

- `/create`
  - POST request
  - With post body that must at least contain a `body` field (plain text), and
    can also contain a `location` field (plain text)
- `/read`
  - GET request
  - By default get's all the updates
- `/update`
  - POST request
  - Required:
    - `?id=<number>` query param
    - Payload with any of:
      - `body`
      - `location`

## Development:

To run the service locally - run `go get` and then `air`
