# PayForward Setup Guide

This guide will help you set up the PayForward application with Keycloak authentication, Neo4j database, and complete development environment.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Keycloak Setup](#keycloak-setup)
- [Social Authentication Setup](#social-authentication-setup)
- [Development Environment](#development-environment)
- [Production Deployment](#production-deployment)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- Docker and Docker Compose
- Git
- (Optional) VS Code with Dev Containers extension

## Quick Start

### 1. Clone the Repository

```bash
git clone <your-repo-url>
cd payforward
```

### 2. Start Services with Docker Compose

```bash
docker-compose up -d
```

This will start:
- **Keycloak** on http://localhost:8180 (admin/admin)
- **Neo4j** on http://localhost:7474 (neo4j/password123)
- **Backend API** on http://localhost:8080
- **Frontend** on http://localhost:3000

### 3. Configure Keycloak

#### Access Keycloak Admin Console

1. Open http://localhost:8180
2. Login with credentials: `admin` / `admin`

#### Import Realm Configuration

The `payforward` realm should be automatically imported. If not:

1. Click "Create Realm" or use the realm dropdown
2. Click "Import"
3. Select `keycloak-realm.json` from the project root
4. Click "Create"

#### Configure Client Secret

1. Select the `payforward` realm
2. Go to **Clients** → **payforward-app**
3. Go to **Credentials** tab
4. Copy the **Client Secret**
5. Update the secret in:
   - `docker-compose.yml` (KEYCLOAK_CLIENT_SECRET)
   - `backend/.env` (if running locally)

### 4. Verify Everything is Running

```bash
# Check backend health
curl http://localhost:8080/api/health

# Check services
docker-compose ps
```

## Keycloak Setup

### User Registration and Login

Keycloak is configured to support:
- **Email/Password authentication**
- **User self-registration**
- **Password reset**
- **Social authentication** (Google, GitHub, Facebook)

### Create Test Users

1. Go to Keycloak Admin Console
2. Select `payforward` realm
3. Go to **Users** → **Add User**
4. Fill in user details:
   - Username: `testuser`
   - Email: `test@example.com`
   - Email Verified: ON
5. Click **Save**
6. Go to **Credentials** tab
7. Set password and disable "Temporary"

### Roles and Permissions

The realm includes three default roles:
- **user**: Regular user (default for all new users)
- **admin**: Administrator with full access
- **moderator**: Moderator with limited admin access

To assign roles to a user:
1. Go to **Users** → Select user
2. Go to **Role Mappings** tab
3. Click **Assign Role**
4. Select the desired role

## Social Authentication Setup

### Google Authentication

1. Create a project in [Google Cloud Console](https://console.cloud.google.com/)
2. Enable Google+ API
3. Create OAuth 2.0 credentials:
   - Application type: Web application
   - Authorized redirect URIs: `http://localhost:8180/realms/payforward/broker/google/endpoint`
4. Copy **Client ID** and **Client Secret**
5. In Keycloak Admin Console:
   - Go to **Identity Providers** → **google**
   - Enable the provider
   - Paste Client ID and Client Secret
   - Save

### GitHub Authentication

1. Go to GitHub Settings → Developer settings → OAuth Apps
2. Click **New OAuth App**
3. Fill in:
   - Application name: PayForward
   - Homepage URL: http://localhost:3000
   - Authorization callback URL: `http://localhost:8180/realms/payforward/broker/github/endpoint`
4. Click **Register application**
5. Generate a new **Client Secret**
6. In Keycloak Admin Console:
   - Go to **Identity Providers** → **github**
   - Enable the provider
   - Paste Client ID and Client Secret
   - Save

### Facebook Authentication

1. Go to [Facebook Developers](https://developers.facebook.com/)
2. Create a new app
3. Add **Facebook Login** product
4. Configure:
   - Valid OAuth Redirect URIs: `http://localhost:8180/realms/payforward/broker/facebook/endpoint`
5. Copy **App ID** and **App Secret**
6. In Keycloak Admin Console:
   - Go to **Identity Providers** → **facebook**
   - Enable the provider
   - Paste Client ID (App ID) and Client Secret (App Secret)
   - Save

## Development Environment

### Using VS Code Dev Containers

1. Install [VS Code](https://code.visualstudio.com/) and the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
2. Open the project in VS Code
3. Click "Reopen in Container" when prompted
4. Wait for the container to build and services to start

### Manual Development Setup

#### Backend

```bash
cd backend

# Install dependencies
go mod download

# Run the server
go run cmd/server/main.go
```

Environment variables (create `.env` file):
```env
PORT=8080
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=password123
JWT_SECRET=change-this-secret-in-production
KEYCLOAK_URL=http://localhost:8180
KEYCLOAK_REALM=payforward
KEYCLOAK_CLIENT_ID=payforward-app
KEYCLOAK_CLIENT_SECRET=your-client-secret-here
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
RATE_LIMIT_PER_MIN=100
ENVIRONMENT=development
```

#### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Run development server
npm run dev
```

Environment variables (create `.env` file):
```env
VITE_API_URL=http://localhost:8080
VITE_KEYCLOAK_URL=http://localhost:8180
VITE_KEYCLOAK_REALM=payforward
VITE_KEYCLOAK_CLIENT_ID=payforward-app
```

### Testing Authentication

#### Using the Frontend

1. Open http://localhost:3000
2. Click "Login" or "Sign Up"
3. You'll be redirected to Keycloak
4. Login with credentials or social provider
5. You'll be redirected back with an access token

#### Using API Directly

Get an access token:

```bash
curl -X POST http://localhost:8180/realms/payforward/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=payforward-app" \
  -d "client_secret=your-client-secret-here" \
  -d "grant_type=password" \
  -d "username=testuser" \
  -d "password=testpassword"
```

Use the token:

```bash
curl http://localhost:8080/api/v1/users/123 \
  -H "Authorization: Bearer <access_token>"
```

## Production Deployment

### Environment Variables

Update these for production:

**Backend:**
- `JWT_SECRET`: Use a strong random secret
- `KEYCLOAK_CLIENT_SECRET`: Use the actual client secret from Keycloak
- `ALLOWED_ORIGINS`: Set to your production domains
- `ENVIRONMENT`: Set to `production`

**Frontend:**
- `VITE_API_URL`: Set to your production API URL
- `VITE_KEYCLOAK_URL`: Set to your production Keycloak URL

### Building for Production

```bash
# Build backend
cd backend
docker build -t payforward-backend:latest .

# Build frontend
cd frontend
docker build -t payforward-frontend:latest --target production .
```

### Using Docker Compose

```bash
# Production deployment
docker-compose -f docker-compose.yml up -d
```

## Troubleshooting

### Keycloak Not Starting

- Check if PostgreSQL is healthy: `docker-compose ps postgres`
- View logs: `docker-compose logs keycloak`
- Ensure port 8180 is not in use

### Backend Cannot Connect to Keycloak

- Verify KEYCLOAK_URL is correct (use container name `keycloak:8080` in Docker)
- Check client secret matches Keycloak configuration
- Ensure Keycloak is fully started before backend

### Token Validation Fails

- Verify client ID and realm name match Keycloak configuration
- Check token expiration time
- Ensure public keys are correctly fetched from Keycloak JWKS endpoint

### CORS Issues

- Add your frontend URL to `ALLOWED_ORIGINS` in backend
- Check Keycloak client configuration has correct redirect URIs

### Database Connection Issues

- Verify Neo4j is running: `docker-compose ps neo4j`
- Check credentials match configuration
- View Neo4j logs: `docker-compose logs neo4j`

## Additional Resources

- [Keycloak Documentation](https://www.keycloak.org/documentation)
- [Neo4j Documentation](https://neo4j.com/docs/)
- [Go Documentation](https://go.dev/doc/)
- [Vite Documentation](https://vitejs.dev/)

## Support

For issues and questions:
- Check the [GitHub Issues](https://github.com/yourusername/payforward/issues)
- Review the logs: `docker-compose logs`
