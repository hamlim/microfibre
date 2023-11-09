# Microfibre

A bare bones Go-based API (using Gin + SQLite) for creating, updating, and
reading status updates.

Think of it as a poor-mans Twitter. No social validation, no fancy features,
just a stream of updates!

## API:

- `/v1/create`
  - POST request
  - With post body that must at least contain a `body` field (plain text), and
    can also contain a `location` field (plain text)
- `/v1/read`
  - GET request
  - By default get's all the updates
- `/v1/update`
  - POST request
  - Required:
    - `?id=<number>` query param
    - Payload with any of:
      - `body`
      - `location`

All calls need to specify both:

- `api-version` request header, at the time of writing this is always `v1`
- A token for making the requests - open an issue if you'd like your client to
  be authorized!

## Development:

To run the service locally - run `go get` and then `air`
