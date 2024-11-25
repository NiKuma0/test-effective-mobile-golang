# Test Project

## Overview

This project demonstrates a simple Go application with PostgreSQL and Docker. It is designed to showcase effective development practices using Docker Compose for containerization, and integrates a PostgreSQL database for data storage.

## Prerequisites

Before running the project, ensure you have the following installed:

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/)
- [Go](https://go.dev/doc/install) (for local developing)

## Running with Docker

1. It is recommended to create a `.env` file inside the `docker/` folder with the following content:

    ```
    COMPOSE_PROJECT_NAME="test-effective-mobile-golang"
    ```

2. To build and start the services:

    ```
    docker compose -f docker/docker-compose.build.yml up
    ```

    1. If you need to specify a custom image, update the `docker/.env` file:

        ```
        IMAGE=<YOUR IMAGE NAME>
        ```

        Then, run the following to start the services:

        ```
        docker compose -f docker/docker-compose.yml up
        ```

## Running in Development Mode

1. Create a `.env` file at the root of the project with the following content:

    ```
    DEBUG=true
    DB=postgres://postgres:postgres@localhost:5432/app
    ```

2. Start the PostgreSQL database:

    ```
    docker compose -f docker/docker-compose.psql.yml up -d
    ```

3. Run the application:

    ```
    go run cmd/main.go
    ```

## Running Tests

To run the tests in this project, use the following Go command:

```
go test ./tests/...
```

## Debugging

A configuration for debugging in Visual Studio Code is already set up. Simply start the PostgreSQL database and press `F5` to start debugging.
