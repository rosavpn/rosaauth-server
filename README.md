# RosaAuth Server

RosaAuth Server is a production-ready Go Fiber web server designed for End-to-End Encrypted (E2EE) 2FA synchronization. It serves as a secure backend for mobile clients to sync encrypted 2FA records and provides an administration interface for user management.

## Features

-   **High Performance:** Built with [Go Fiber (v2)](https://gofiber.io/).
-   **Security:**
    -   JWT-based authentication for both API and Admin UI.
    -   "Dumb Pipe" architecture: The server stores encrypted blobs and does not know the master key.
-   **Database:** PostgreSQL with standardized SQL migrations.
-   **Admin UI:** Embedded Single Page Application (SPA) for managing users.
-   **Observability:** Structured JSON logging via `zerolog`.

## Tech Stack

-   **Language:** Go (Golang)
-   **Framework:** Fiber v2
-   **Database:** PostgreSQL (`lib/pq`)
-   **Logging:** Zerolog
-   **Frontend:** HTML5, Alpine.js, TailwindCSS

### Docker Support

To run the application and database together, it's best to run them in the same Docker network.

1.  **Create a Network**
    ```bash
    docker network create rosaauth-net
    ```

2.  **Run PostgreSQL**
    ```bash
    docker run -d \
      --name rosaauth-db \
      --network rosaauth-net \
      -e POSTGRES_USER=user \
      -e POSTGRES_PASSWORD=password \
      -e POSTGRES_DB=rosaauth \
      postgres:18-alpine
    ```

3.  **Run RosaAuth Server**
    ```bash
    docker run --rm \
      --name rosaauth-server \
      --network rosaauth-net \
      -p 3000:8080 \
      -e DB_URL="postgres://user:password@rosaauth-db:5432/rosaauth?sslmode=disable" \
      -e JWT_SECRET="your-secure-secret-key" \
      -e ADMIN_EMAIL="admin@example.com" \
      -e ADMIN_PASSWORD="securepassword" \
      -e SALT_KEY_STRING="some-random-string" \
      ghcr.io/brsyuksel/rosaauth-server:latest
    ```

    *Note: If building locally, replace the image name with `rosaauth-server`.*

## Getting Started

### Prerequisites

-   Go 1.23 or higher
-   PostgreSQL

### Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/brsyuksel/rosaauth-server.git
    cd rosaauth-server
    ```

3.  Set up the database:

    Since we are running locally in this step, ensure you have a Postgres database running.
    ```bash
    # Example using Docker for just the DB (mapped to localhost)
    docker run -d -e POSTGRES_USER=user -e POSTGRES_PASSWORD=password -e POSTGRES_DB=rosaauth -p 5432:5432 postgres:18-alpine
    ```

4.  Configure Environment Variables:
    Create a `.env` file or export them directly:
    ```bash
    export PORT=3000
    export DB_URL="postgres://user:password@localhost:5432/rosaauth?sslmode=disable"
    export JWT_SECRET="your-secure-secret-key"
    export ADMIN_EMAIL="admin@example.com"
    export ADMIN_PASSWORD="securepassword"
    export SALT_KEY_STRING="some-random-string"
    export LOG_LEVEL="info"
    ```

4.  Run the server:
    ```bash
    go mod tidy
    go run cmd/server/main.go
    ```
    The server will automatically run migrations on startup and ensure the Admin user exists.

### Usage

#### Admin UI
Open your browser and navigate to `http://localhost:3000`. Log in using the `ADMIN_EMAIL` and `ADMIN_PASSWORD` you configured.

#### Client API

**1. Login**
-   **Endpoint:** `POST /api/v1/login`
-   **Body:** `{"email": "user@example.com", "password": "password"}`
-   **Response:** `{"token": "jwt-token"}`

**2. Sync Records**
-   **Endpoint:** `POST /api/v1/sync`
-   **Headers:** `Authorization: Bearer <token>`
-   **Body:** Array of operations
    ```json
    [
      {
        "op": "upsert",
        "data": {
          "id": "uuid-v4",
          "encrypted_data": "base64_string_given_by_user"
        }
      },
      {
        "op": "delete",
        "data": { "id": "uuid-v4-to-delete" }
      }
    ]
    ```
-   **Response:** Array of current records: `[{"id": "...", "encrypted_data": "..."}]`

## License

This project is licensed under the MIT License
