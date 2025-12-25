# PayForward - Quick Start Guide

## 🚀 Get Up and Running in 5 Minutes

### Step 1: Start All Services

```bash
docker-compose up -d
```

Wait 30-60 seconds for all services to start.

### Step 2: Verify Services Are Running

```bash
# Check all services
docker-compose ps

# You should see:
# - postgres (healthy)
# - keycloak (healthy)
# - neo4j (healthy)
# - backend (running)
# - frontend (running)
```

### Step 3: Access the Application

Open your browser and navigate to:

| Service | URL | Credentials |
|---------|-----|-------------|
| **Frontend** | http://localhost:3000 | - |
| **Backend API** | http://localhost:8080/api/health | - |
| **Keycloak Admin** | http://localhost:8180 | admin / admin |
| **Neo4j Browser** | http://localhost:7474 | neo4j / password123 |

### Step 4: Configure Keycloak Client Secret

1. Open http://localhost:8180
2. Login with `admin` / `admin`
3. Select the `payforward` realm (should auto-import)
4. Go to **Clients** → **payforward-app**
5. Go to **Credentials** tab
6. Copy the **Client Secret**
7. Update the backend:
   ```bash
   # Edit docker-compose.yml and update KEYCLOAK_CLIENT_SECRET
   # Then restart backend
   docker-compose restart backend
   ```

### Step 5: Test Authentication

#### Create a Test User

1. In Keycloak Admin Console (still logged in)
2. Go to **Users** → **Add User**
3. Fill in:
   - Username: `testuser`
   - Email: `test@example.com`
   - Email Verified: **ON**
4. Click **Save**
5. Go to **Credentials** tab
6. Set password: `testpass123`
7. Disable "Temporary"
8. Click **Set Password**

#### Test Login via API

```bash
# Get access token
curl -X POST http://localhost:8180/realms/payforward/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=payforward-app" \
  -d "client_secret=YOUR_CLIENT_SECRET_HERE" \
  -d "grant_type=password" \
  -d "username=testuser" \
  -d "password=testpass123" | jq

# Copy the access_token from the response
# Use it to call protected API endpoints

curl http://localhost:8080/api/v1/users/123 \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" | jq
```

## 🎯 Next Steps

### Enable Social Authentication

#### Google Login

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable **Google+ API**
4. Go to **Credentials** → **Create Credentials** → **OAuth 2.0 Client ID**
5. Application type: **Web application**
6. Add authorized redirect URI:
   ```
   http://localhost:8180/realms/payforward/broker/google/endpoint
   ```
7. Copy **Client ID** and **Client Secret**
8. In Keycloak Admin Console:
   - Go to **Identity Providers**
   - Find **google** provider
   - Click to configure
   - Enable the provider
   - Paste Client ID and Client Secret
   - Save

#### GitHub Login

1. Go to GitHub Settings → Developer settings → OAuth Apps
2. Click **New OAuth App**
3. Fill in:
   - Application name: `PayForward Local`
   - Homepage URL: `http://localhost:3000`
   - Authorization callback URL:
     ```
     http://localhost:8180/realms/payforward/broker/github/endpoint
     ```
4. Register application
5. Generate **Client Secret**
6. In Keycloak Admin Console:
   - Go to **Identity Providers**
   - Find **github** provider
   - Enable and configure with Client ID and Secret

#### Facebook Login

1. Go to [Facebook Developers](https://developers.facebook.com/)
2. Create a new app
3. Add **Facebook Login** product
4. Settings → Basic: Copy App ID and App Secret
5. Facebook Login → Settings:
   - Add Valid OAuth Redirect URIs:
     ```
     http://localhost:8180/realms/payforward/broker/facebook/endpoint
     ```
6. In Keycloak Admin Console:
   - Go to **Identity Providers**
   - Find **facebook** provider
   - Enable and configure

### Development

#### Backend Development

```bash
cd backend

# Install dependencies
go mod download

# Run locally (outside Docker)
export NEO4J_URI=bolt://localhost:7687
export NEO4J_USER=neo4j
export NEO4J_PASSWORD=password123
export KEYCLOAK_URL=http://localhost:8180
export KEYCLOAK_REALM=payforward
export KEYCLOAK_CLIENT_ID=payforward-app
export KEYCLOAK_CLIENT_SECRET=your-secret-here

go run cmd/server/main.go
```

#### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Create .env file
cat > .env <<EOF
VITE_API_URL=http://localhost:8080
VITE_KEYCLOAK_URL=http://localhost:8180
VITE_KEYCLOAK_REALM=payforward
VITE_KEYCLOAK_CLIENT_ID=payforward-app
EOF

# Run dev server
npm run dev
```

## 🔧 Using Make Commands

```bash
# Show all available commands
make help

# Start services
make up

# Stop services
make down

# View logs
make logs

# View specific service logs
make logs-backend
make logs-frontend
make logs-keycloak
make logs-neo4j

# Install dependencies
make install

# Run tests
make test

# Build everything
make build

# Open Keycloak admin console
make keycloak-setup

# Open Neo4j browser
make neo4j-browser
```

## 🐛 Common Issues

### Services Won't Start

```bash
# Check what's using the ports
sudo lsof -i :3000  # Frontend
sudo lsof -i :8080  # Backend
sudo lsof -i :8180  # Keycloak
sudo lsof -i :7474  # Neo4j HTTP
sudo lsof -i :7687  # Neo4j Bolt

# Clean everything and restart
make clean
make up
```

### Keycloak Not Ready

```bash
# Keycloak takes 30-60 seconds to start
# Check logs
docker-compose logs keycloak

# Wait for this message:
# "Keycloak ... started"
```

### Backend Can't Connect to Keycloak

```bash
# Ensure Keycloak is fully started
docker-compose ps

# Check backend logs
docker-compose logs backend

# Restart backend after Keycloak is ready
docker-compose restart backend
```

### Token Validation Fails

1. Verify client secret is correct
2. Check realm name is `payforward`
3. Ensure client ID is `payforward-app`
4. Verify token hasn't expired (default 1 hour)

## 📚 Additional Resources

- [Full Setup Guide](SETUP.md) - Detailed configuration instructions
- [README](README.md) - Complete project documentation
- [Keycloak Docs](https://www.keycloak.org/documentation)
- [Neo4j Docs](https://neo4j.com/docs/)

## ✅ Verification Checklist

- [ ] All services are running (`docker-compose ps`)
- [ ] Backend health check passes (`curl http://localhost:8080/api/health`)
- [ ] Keycloak admin console is accessible
- [ ] Neo4j browser is accessible
- [ ] Client secret is configured
- [ ] Test user can login and get token
- [ ] Protected endpoints work with token

**You're all set! Happy coding! 🎉**
