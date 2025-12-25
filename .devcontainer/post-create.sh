#!/bin/bash

set -e

echo "🚀 Running post-create setup..."

# Install backend dependencies
echo "📦 Installing backend dependencies..."
cd /workspace/backend
go mod download
go mod tidy

# Install frontend dependencies
echo "📦 Installing frontend dependencies..."
cd /workspace/frontend
npm install

# Create .env file if it doesn't exist
if [ ! -f /workspace/.env ]; then
    echo "📝 Creating .env file..."
    cat > /workspace/.env <<EOF
# Backend
PORT=8080
NEO4J_URI=bolt://neo4j:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=password123
JWT_SECRET=change-this-secret-in-production
KEYCLOAK_URL=http://keycloak:8080
KEYCLOAK_REALM=payforward
KEYCLOAK_CLIENT_ID=payforward-app
KEYCLOAK_CLIENT_SECRET=your-client-secret-here
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
RATE_LIMIT_PER_MIN=100
ENVIRONMENT=development

# Frontend
VITE_API_URL=http://localhost:8080
VITE_KEYCLOAK_URL=http://localhost:8180
VITE_KEYCLOAK_REALM=payforward
VITE_KEYCLOAK_CLIENT_ID=payforward-app
EOF
fi

# Set permissions
chmod +x /workspace/.devcontainer/post-create.sh

echo "✅ Post-create setup complete!"
echo ""
echo "📌 Next steps:"
echo "  1. Run 'docker-compose up -d' to start services"
echo "  2. Access Keycloak admin at http://localhost:8180 (admin/admin)"
echo "  3. Access Neo4j browser at http://localhost:7474 (neo4j/password123)"
echo "  4. Start backend: cd backend && go run cmd/server/main.go"
echo "  5. Start frontend: cd frontend && npm run dev"
echo ""
