# Crowfather

Crowfather is a small Go service that exposes HTTP endpoints for interacting with GroupMe and OpenAI.  The server is configured using environment variables and uses the [gin](https://github.com/gin-gonic/gin) framework for routing.

## Structure

```
internal/
  config       - loading configuration from environment variables
  groupme      - send messages to the GroupMe API
  handlers/
      message_handler   - process incoming GroupMe messages
      meltdown_handler  - simple text handler
      test_handler      - simple text handler
  open_ai      - wrapper around the OpenAI API
  router       - HTTP routes and middleware
  main.go      - program entry point
```

## Endpoints

* `GET /ping` – health check returning `pong`.
* `POST /message` – receives a GroupMe webhook payload and responds using OpenAI.
* `POST /meltdown` – send a single message to OpenAI.
* `POST /test` – test endpoint protected by the `API_KEY` header or query parameter.

## Configuration

The service relies on several environment variables:

```
OPENAI_API_KEY          OpenAI API key
GROUPME_BOT_ID          GroupMe bot identifier
GROUPME_BOT_TOKEN       authentication token for GroupMe
GROUPME_ASSISTANT_ID    OpenAI assistant ID for group messages
MELTDOWN_ASSISTANT_ID   OpenAI assistant ID for meltdown messages
TEST_ASSISTANT_ID       OpenAI assistant ID for test messages
API_KEY                 simple API key used by the test route
DB_USER                 database username
DB_PASS                 database password
DB_HOST                 database host
DB_NAME                 database name
```

Set these variables before running the server with `go run ./internal`.

## Running Tests

Unit tests cover the configuration loader, the GroupMe client and the database layer. Execute them with:

```
go test ./...
```

