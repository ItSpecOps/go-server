# Chirpy Go Server

A simple HTTP server for the Chirpy project, written in Go.

## Features

- Health check endpoint: `/api/healthz`
- Metrics endpoint: `/admin/metrics`
- Reset metrics endpoint: `/admin/reset`
- Chirp validation endpoint: `/api/validate_chirp`
- Static file server at `/app/`
- In-memory request counting

## Requirements

- Go 1.25+
- PostgreSQL (for future database features)

## Setup

1. Clone the repository.
2. Copy `.env` and set your database URL if needed.
3. Build and run the server:

   ```sh
   go build -o out
   ./out
   ```

4. Visit [http://localhost:8080/app/](http://localhost:8080/app/) to see the static site.

## API Endpoints

- `GET /api/healthz`  
  Health check. Returns `OK`.

- `GET /admin/metrics`  
  Returns the number of visits as HTML.

- `POST /admin/reset`  
  Resets the visit counter.

- `POST /api/validate_chirp`  
  Accepts JSON:  
  ```json
  { "body": "your chirp here" }
  ```
  Returns cleaned chirp or error if too long or invalid.

## License

MIT
