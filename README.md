🎥 ByteStream Playback API

A minimal, idiomatic Go service that resolves a playable video URL based on:

👤 User identity (standard vs premium)

📅 Content availability window

🎬 Video metadata lookup

The project is intentionally structured to keep business logic isolated and testable, with a simple local development workflow.

📁 Project Structure
cmd/
  playback-api/        # Main HTTP API
  mock-upstreams/      # Local Identity & Availability mocks (dev only)

internal/
  api/                 # HTTP handlers & middleware
  catalog/             # Hardcoded video metadata
  clients/             # HTTP clients for upstream services
  config/              # Environment-based configuration
  domain/              # Core business logic + unit tests

.env.example           # Example runtime configuration
Makefile               # Common development commands
go.mod
README.md

✅ Prerequisites

Go 1.22+

No external services required (local mocks provided)

⚙️ Configuration

An example environment file is provided at the repository root:

.env.example


To create a local environment file:

cp .env.example .env
source .env


⚠️ Do not commit .env files — they are ignored by .gitignore.

🛠 Using the Makefile (Recommended)

The Makefile provides a simple, consistent command interface for common workflows.
It reduces setup friction and acts as executable documentation.

Available commands
make test          # Run all tests
make test-domain   # Run only domain (business logic) tests
make mocks         # Start mock identity & availability services
make run           # Start the playback API


Using the Makefile is optional, but strongly recommended.

🧪 Running Domain Unit Tests (Only)

The core business logic lives in:

internal/domain/domain.go


Unit tests live alongside it:

internal/domain/domain_test.go

Run only domain tests
make test-domain


Or directly with Go:

go test ./internal/domain

Verbose output
go test -v ./internal/domain

With coverage
go test ./internal/domain -cover


✅ These tests do not start the HTTP server and do not require any upstream services.

🔌 Running the Mock Upstream Services

A local mock server simulates:

GET /identity/userinfo

GET /availability/availabilityinfo/{video_id}

Start mocks
make mocks


Or directly:

go run ./cmd/mock-upstreams


Mocks listen on:

http://127.0.0.1:9001

Verify mocks
curl http://127.0.0.1:9001/identity/userinfo
curl http://127.0.0.1:9001/availability/availabilityinfo/46325

▶️ Running the Playback API
1️⃣ Set environment variables
export IDENTITY_BASE_URL="http://127.0.0.1:9001"
export AVAILABILITY_BASE_URL="http://127.0.0.1:9001"
export S3_PLAYBACK_BASEURL="https://s3.eu-west-1.amazonaws.com/bytestreamfake"
export PORT="8080"
export HTTP_TIMEOUT="3s"

2️⃣ Start the API
make run


Or directly:

go run ./cmd/playback-api


The service runs on:

http://localhost:8080

🔍 Testing the API End-to-End
Health checks
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz

Playback request
curl -H "Authorization: bearer testtoken" \
     http://localhost:8080/playback/46325

Example response (premium user)
{
  "video_id": 46325,
  "title": "Example Video 001",
  "playback_baseurl": "https://s3.eu-west-1.amazonaws.com/bytestreamfake",
  "playback_filename": "example001-premium",
  "playback_extension": ".mp4"
}

📝 Notes

Playback URLs are intentionally unprotected (static S3) for early frontend integration.

Video metadata is currently hardcoded; a lookup service is expected in a future iteration.

Business logic is fully unit-tested and independent of HTTP and infrastructure.