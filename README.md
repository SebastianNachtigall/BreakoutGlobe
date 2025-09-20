# BreakoutGlobe

An innovative online workshop platform that combines an interactive world map with immersive video/audio breakout rooms.

## Features

- Interactive world map with avatar-based user presence
- Real-time Points of Interest (POI) system
- WebSocket-based real-time communication
- No authentication required - instant access
- Docker-first development and deployment

## Development Setup

### Prerequisites

- Docker and Docker Compose
- Node.js 18+ (for local development)
- Go 1.21+ (for local development)

### Quick Start

1. Clone the repository:
```bash
git clone <repository-url>
cd breakoutglobe
```

2. Start the development environment:
```bash
docker compose up
```

3. Access the application:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Health check: http://localhost:8080/health

### Development Workflow

The project follows Test-Driven Development (TDD) methodology:

1. Write failing tests first (Red)
2. Write minimal code to make tests pass (Green)
3. Refactor while keeping tests green (Refactor)

### Project Structure

```
├── backend/                 # Go backend service
│   ├── cmd/server/         # Application entry point
│   ├── internal/           # Internal packages
│   ├── migrations/         # Database migrations
│   └── Dockerfile          # Backend container
├── frontend/               # React TypeScript frontend
│   ├── src/               # Source code
│   ├── public/            # Static assets
│   └── Dockerfile.dev     # Frontend development container
├── .github/workflows/     # CI/CD pipelines
└── compose.yml            # Development environment
```

### Testing

Run tests locally:

```bash
# Frontend tests
cd frontend
npm run test

# Backend tests
cd backend
go test ./...

# E2E tests
cd frontend
npm run test:e2e
```

### Contributing

1. All features must be implemented using TDD
2. Tests must pass before committing
3. Follow the established project structure
4. Use conventional commit messages

## Architecture

- **Frontend**: React 18 + TypeScript + Vite + Tailwind CSS
- **Backend**: Go + Gin framework + WebSockets
- **Database**: PostgreSQL for persistent data
- **Cache**: Redis for sessions and real-time features
- **Infrastructure**: Docker containers with health checks

## License

[License information to be added]