# PayForward Backend

A Go-based backend service for the PayForward application, built with Neo4j graph database.

## Features

- RESTful API with Go standard library http package
- Neo4j graph database integration
- JWT authentication
- Rate limiting and security middleware
- Comprehensive test coverage
- Docker support

## Project Structure

```
backend/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── auth/            # Authentication logic (Keycloak)
│   ├── database/        # Database client and interfaces
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # HTTP middleware (CORS, auth, logging, etc.)
│   └── models/          # Data models and types
├── Makefile             # Build and test automation
├── Dockerfile           # Docker image configuration
└── go.mod               # Go module dependencies
```

## Prerequisites

- Go 1.24 or higher
- Neo4j 5.15 or higher
- Docker (optional, for containerized deployment)
- Make (for using Makefile commands)

## Getting Started

### 1. Install Dependencies

```bash
make install-deps
```

### 2. Set Up Environment Variables

Create a `.env` file in the backend directory:

```env
PORT=8080
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=password
JWT_SECRET=your-secret-key-change-in-production
ENVIRONMENT=development
ALLOWED_ORIGINS=*
RATE_LIMIT_PER_MIN=100

# Optional: Keycloak Configuration
KEYCLOAK_URL=
KEYCLOAK_REALM=
KEYCLOAK_CLIENT_ID=
KEYCLOAK_CLIENT_SECRET=
```

### 3. Build the Application

```bash
make build
```

The binary will be created in `bin/server`.

### 4. Run the Application

```bash
make run
```

Or run the binary directly:

```bash
./bin/server
```

## Testing

### Run All Tests

```bash
make test
```

### Run Tests with Verbose Output

```bash
make test-verbose
```

### Run Short Tests (Skip Integration Tests)

```bash
make test-short
```

### Run Tests with Coverage Report

```bash
make test-coverage
```

This generates `coverage.html` that you can open in your browser.

### Run Benchmarks

```bash
make benchmark
```

## Test Coverage

Current test coverage by package:

- **models**: 100% coverage - Full unit test coverage for all data models
- **middleware**: 63.3% coverage - Comprehensive middleware testing (CORS, auth, rate limiting, security headers)
- **handlers**: 10.3% coverage - Basic handler tests (uses mocks for database)
- **database**: Integration tests using testcontainers

### Test Types

1. **Unit Tests**: Test individual functions and methods in isolation
   - Location: `*_test.go` files alongside source code
   - Run with: `make test-short`

2. **Integration Tests**: Test database operations with real Neo4j instance
   - Location: `internal/database/neo4j_integration_test.go`
   - Uses testcontainers to spin up Neo4j in Docker
   - Run with: `make test` (skipped in short mode)

## Available Make Commands

```bash
make help              # Display all available commands
make build             # Build the application
make run               # Run the application
make test              # Run all tests
make test-verbose      # Run tests with verbose output
make test-coverage     # Run tests and generate coverage report
make test-short        # Run tests excluding integration tests
make benchmark         # Run benchmarks
make clean             # Clean build artifacts
make fmt               # Format code
make vet               # Run go vet
make lint              # Run golangci-lint
make check             # Run fmt, vet, lint, and test
make docker-build      # Build Docker image
make docker-up         # Start Docker Compose services
make docker-down       # Stop Docker Compose services
make dev               # Run with hot reload (requires air)
```

## Code Quality

### Format Code

```bash
make fmt
```

### Run Static Analysis

```bash
make vet
```

### Run Linter

```bash
make lint
```

### Run All Checks

```bash
make check
```

This runs formatting, vetting, linting, and all tests.

## API Endpoints

### Health Check
- `GET /api/health` - Check service health

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/logout` - Logout user
- `POST /api/v1/auth/refresh` - Refresh authentication token

### Users
- `GET /api/v1/users/{id}` - Get user by ID
- `POST /api/v1/users` - Create new user
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user

### Acts of Kindness
- `GET /api/v1/acts` - List all acts (paginated)
- `POST /api/v1/acts` - Create new act
- `GET /api/v1/acts/{id}` - Get act by ID
- `PUT /api/v1/acts/{id}` - Update act
- `DELETE /api/v1/acts/{id}` - Delete act

### Chains
- `GET /api/v1/chains/{id}` - Get chain by ID
- `GET /api/v1/users/{id}/chains` - Get chains for user

### Statistics
- `GET /api/v1/stats/global` - Get global statistics
- `GET /api/v1/stats/user/{id}` - Get user statistics

### Testimonials
- `GET /api/v1/testimonials` - List approved testimonials
- `POST /api/v1/testimonials` - Create new testimonial

## Development

### Hot Reload (Development Mode)

```bash
make dev
```

This uses [air](https://github.com/cosmtrek/air) for hot reloading during development.

### Docker Development

Start all services (backend + Neo4j):

```bash
make docker-up
```

Stop all services:

```bash
make docker-down
```

View logs:

```bash
make docker-logs
```

## Production Deployment

### Build Docker Image

```bash
make docker-build
```

### Environment Configuration

Ensure all production environment variables are properly set:
- Use strong JWT secrets
- Configure proper CORS origins
- Set appropriate rate limits
- Use secure database credentials

## Contributing

1. Write tests for new features
2. Ensure all tests pass: `make check`
3. Format code: `make fmt`
4. Run linter: `make lint`

## License

MIT
