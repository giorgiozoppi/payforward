# PayForward

A pay-it-forward platform with Keycloak authentication, Neo4j graph database, and modern web technologies.

## Features

- 🔐 **Keycloak Authentication** - Support for email/password and social logins (Google, GitHub, Facebook)
- 🌐 **Graph Database** - Neo4j for tracking pay-it-forward chains and relationships
- 🚀 **Modern Stack** - Go backend, React frontend, Docker containerization
- 🔒 **Security** - JWT tokens, CORS, rate limiting, security headers
- 📦 **Dev Containers** - Full development environment with VS Code
- 🤖 **CI/CD** - GitHub Actions for automated testing and deployment

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Start all services
docker-compose up -d

# Services will be available at:
# - Frontend: http://localhost:3000
# - Backend API: http://localhost:8080
# - Keycloak: http://localhost:8180 (admin/admin)
# - Neo4j Browser: http://localhost:7474 (neo4j/password123)
```

### Using Make

```bash
# Install dependencies
make install

# Start all services
make up

# View logs
make logs

# Stop services
make down

# See all commands
make help
```

### Using VS Code Dev Containers

1. Install [VS Code](https://code.visualstudio.com/) and the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
2. Open the project in VS Code
3. Click "Reopen in Container" when prompted
4. All dependencies and services will be automatically configured

## Architecture

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   Frontend  │──────│  Backend    │──────│   Neo4j     │
│   (React)   │      │    (Go)     │      │  Database   │
└─────────────┘      └─────────────┘      └─────────────┘
       │                     │
       │             ┌───────▼──────┐
       └─────────────│   Keycloak   │
                     │    (Auth)    │
                     └──────────────┘
```

## Authentication Setup

### Keycloak Configuration

1. **Access Keycloak Admin Console**
   - URL: http://localhost:8180
   - Credentials: `admin` / `admin`

2. **The `payforward` realm is auto-imported**
   - If not, import `keycloak-realm.json`

3. **Get Client Secret**
   - Go to Clients → payforward-app → Credentials
   - Copy the client secret
   - Update in `docker-compose.yml` and backend `.env`

### Social Authentication

Configure social providers in Keycloak:

**Google:**
1. Create OAuth credentials in [Google Cloud Console](https://console.cloud.google.com/)
2. Add redirect URI: `http://localhost:8180/realms/payforward/broker/google/endpoint`
3. Add Client ID and Secret in Keycloak → Identity Providers → google

**GitHub:**
1. Create OAuth App in GitHub Settings → Developer settings
2. Add callback URL: `http://localhost:8180/realms/payforward/broker/github/endpoint`
3. Add Client ID and Secret in Keycloak → Identity Providers → github

**Facebook:**
1. Create app in [Facebook Developers](https://developers.facebook.com/)
2. Add redirect URI: `http://localhost:8180/realms/payforward/broker/facebook/endpoint`
3. Add App ID and Secret in Keycloak → Identity Providers → facebook

See [SETUP.md](SETUP.md) for detailed instructions.

## Development

### Backend (Go)

```bash
cd backend

# Install dependencies
go mod download

# Run server
go run cmd/server/main.go

# Run tests
go test -v ./...

# Build
go build -o server ./cmd/server
```

### Frontend (React + Vite)

```bash
cd frontend

# Install dependencies
npm install

# Run dev server
npm run dev

# Build for production
npm run build

# Run linter
npm run lint
```

## Project Structure

```
payforward/
├── backend/
│   ├── cmd/server/          # Main application entry
│   ├── internal/
│   │   ├── auth/           # Keycloak authentication
│   │   ├── database/       # Neo4j connection
│   │   ├── handlers/       # HTTP handlers
│   │   ├── middleware/     # HTTP middleware
│   │   └── models/         # Data models
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/
│   ├── public/
│   ├── Dockerfile
│   ├── nginx.conf
│   └── package.json
├── .devcontainer/          # VS Code dev container config
├── .github/workflows/      # CI/CD pipelines
├── k8s/                    # Kubernetes manifests
├── docker-compose.yml      # Production compose
├── docker-compose.dev.yml  # Development compose
├── keycloak-realm.json     # Keycloak realm config
├── Makefile               # Development shortcuts
└── README.md
```

## Environment Variables

Copy `.env.example` to `.env` and configure:

```env
# Backend
KEYCLOAK_URL=http://keycloak:8080
KEYCLOAK_REALM=payforward
KEYCLOAK_CLIENT_ID=payforward-app
KEYCLOAK_CLIENT_SECRET=your-secret-here

# Frontend
VITE_API_URL=http://localhost:8080
VITE_KEYCLOAK_URL=http://localhost:8180
VITE_KEYCLOAK_REALM=payforward
```

## API Endpoints

### Public Endpoints
- `GET /api/health` - Health check
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login

### Protected Endpoints (Require Authentication)
- `GET /api/v1/users/{id}` - Get user profile
- `PUT /api/v1/users/{id}` - Update user profile
- `GET /api/v1/acts` - List acts
- `POST /api/v1/acts` - Create act
- `GET /api/v1/chains/{id}` - Get chain details
- `GET /api/v1/stats/global` - Global statistics

## Testing

```bash
# Run all tests
make test

# Backend tests
make test-backend

# Frontend tests
make test-frontend

# Linting
make lint
```

## Deployment

### Docker

```bash
# Build images
docker-compose build

# Deploy
docker-compose up -d
```

### Kubernetes

```bash
# Apply manifests
kubectl apply -f k8s/
```

## CI/CD

GitHub Actions workflows are configured for:
- **Backend CI/CD**: Build, test, security scan, Docker image
- **Frontend CI/CD**: Build, test, lint, security scan, Docker image

Workflows run on push to `main` and `develop` branches and on pull requests.

## Monitoring and Observability

- **Backend Health**: http://localhost:8080/api/health
- **Keycloak Metrics**: http://localhost:8180/metrics
- **Neo4j Metrics**: Available through Neo4j Browser

## Troubleshooting

### Services Not Starting

```bash
# Check service status
docker-compose ps

# View logs
docker-compose logs [service-name]

# Restart services
docker-compose restart
```

### Authentication Issues

1. Verify Keycloak is running: `curl http://localhost:8180/health`
2. Check client secret matches between Keycloak and backend
3. Verify redirect URIs in Keycloak client configuration

### Database Connection Issues

```bash
# Test Neo4j connection
docker-compose exec neo4j cypher-shell -u neo4j -p password123

# Check Neo4j logs
docker-compose logs neo4j
```

See [SETUP.md](SETUP.md) for more troubleshooting tips.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.

## Support

- 📖 [Setup Guide](SETUP.md)
- 🐛 [Report Issues](https://github.com/yourusername/payforward/issues)
- 💬 [Discussions](https://github.com/yourusername/payforward/discussions)
