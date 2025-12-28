# BabyTrack

[![CI](https://github.com/ninenine/babytrack/actions/workflows/ci.yml/badge.svg)](https://github.com/ninenine/babytrack/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/ninenine/babytrack)](https://github.com/ninenine/babytrack/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ninenine/babytrack)](https://go.dev/)
[![License](https://img.shields.io/github/license/ninenine/babytrack)](LICENSE)

A mobile-first web application for tracking baby care activities including feeding, sleep, medications, vaccinations, appointments, and notes. Built with offline-first architecture for reliable use even without internet connectivity.

## Features

- **Feeding Tracking** - Log breast, bottle, formula, and solid food feedings with duration and amounts
- **Sleep Tracking** - Track naps and night sleep with active session timers
- **Medication Management** - Manage medications with dosage schedules and dose logging
- **Vaccination Records** - Track vaccination schedules with auto-generated CDC recommendations
- **Appointments** - Schedule and manage doctor visits, checkups, and specialist appointments
- **Notes** - Keep pinned notes and observations about your child
- **Timeline View** - See all activities in a chronological feed
- **Multi-child Support** - Switch between multiple children in one family
- **Offline-first** - Works without internet, syncs when back online
- **Dark Mode** - Full dark mode support

## Tech Stack

### Backend
- **Go 1.25** - Server runtime
- **Gin** - HTTP web framework
- **PostgreSQL** - Primary database
- **golang-migrate** - Database migrations
- **JWT** - Authentication tokens

### Frontend
- **React 19** - UI framework
- **React Router 7** - Client-side routing
- **TypeScript** - Type safety
- **Tailwind CSS 4** - Styling
- **shadcn/ui** - Component library (Radix UI primitives)
- **TanStack Query** - Server state management
- **Zustand** - Client state management
- **Dexie** - IndexedDB wrapper for offline storage
- **Vite** - Build tool

## Project Structure

```
babytrack/
├── cmd/
│   └── server/          # Application entrypoint
├── configs/             # Configuration files
├── internal/
│   ├── app/             # HTTP server, router, handlers
│   ├── auth/            # Authentication (Google OAuth, JWT)
│   ├── db/              # Database connection and migrations
│   ├── family/          # Family and child management
│   ├── feeding/         # Feeding tracking
│   ├── sleep/           # Sleep tracking
│   ├── medication/      # Medication management
│   ├── vaccination/     # Vaccination records
│   ├── appointment/     # Appointment scheduling
│   ├── notes/           # Notes feature
│   ├── jobs/            # Background jobs
│   └── sync/            # Offline sync service
└── web/                 # React frontend
    ├── src/
    │   ├── app/         # App providers and routes
    │   ├── components/  # React components
    │   │   ├── ui/      # shadcn/ui components
    │   │   ├── layout/  # App shell, nav, etc.
    │   │   └── [feature]/
    │   ├── db/          # Dexie schema and sync
    │   ├── hooks/       # Custom hooks and queries
    │   ├── lib/         # Utilities
    │   ├── pages/       # Route pages
    │   ├── stores/      # Zustand stores
    │   └── types/       # TypeScript types
    └── package.json
```

## Getting Started

### Prerequisites

- Go 1.25+
- Node.js 20+
- pnpm
- Docker (for PostgreSQL)
- pre-commit (`brew install pre-commit`)
- Security tools (`brew install gitleaks gosec trivy`)

### Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/ninenine/babytrack.git
   cd babytrack
   ```

2. **Install all dependencies and set up hooks**
   ```bash
   make install
   ```

3. **Start the database**
   ```bash
   make db-up
   ```

4. **Run database migrations**
   ```bash
   make migrate
   ```

5. **Start development server**
   ```bash
   make dev
   ```

   The app will be available at `http://localhost:8080`

### Development

For frontend development with hot reload:

```bash
# Terminal 1: Start the backend
make dev

# Terminal 2: Start Vite dev server
cd web && pnpm dev
```

The Vite dev server runs on `http://localhost:5173` and proxies API requests to the Go backend.

## Available Commands

### Development
| Command | Description |
|---------|-------------|
| `make install` | Install all dependencies (Go + web) |
| `make dev` | Start database and run server in development mode |
| `make dev-web` | Start Vite dev server with hot reload |

### Database
| Command | Description |
|---------|-------------|
| `make db-up` | Start PostgreSQL container |
| `make db-down` | Stop PostgreSQL container |
| `make db-reset` | Reset database (drop and recreate) |
| `make migrate` | Run database migrations |

### Build
| Command | Description |
|---------|-------------|
| `make build` | Build web UI and server binary |
| `make build-web` | Build only the web UI |
| `make build-server` | Build web UI and server binary |
| `make run` | Run the built binary |

### Code Quality
| Command | Description |
|---------|-------------|
| `make lint` | Run linters (Go vet + ESLint) |
| `make format` | Format all code (gofmt + Prettier) |
| `make pre-commit` | Run all pre-commit hooks |
| `make test` | Run Go tests |

### Other
| Command | Description |
|---------|-------------|
| `make clean` | Clean build artifacts |

## Code Quality

### Pre-commit Hooks

This project uses [pre-commit](https://pre-commit.com/) for automated code quality checks. Hooks are installed automatically via `make install`.

| Hook | Purpose |
|------|---------|
| `trailing-whitespace` | Remove trailing whitespace |
| `end-of-file-fixer` | Ensure files end with newline |
| `check-yaml` | Validate YAML syntax |
| `check-added-large-files` | Block files > 1MB |
| `detect-private-key` | Block commits with private keys |
| `detect-aws-credentials` | Block commits with AWS credentials |
| `gitleaks` | Scan for secrets (API keys, tokens) |
| `gosec` | Go security static analysis |
| `go-fmt` | Format Go code |
| `go-vet` | Go static analysis |
| `go-mod-tidy` | Keep go.mod clean |
| `eslint` | Lint TypeScript/React code |
| `prettier` | Format CSS files |
| `commitlint` | Enforce conventional commits |

### Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope?): subject

body?

footer?
```

**Allowed types:** `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

**Examples:**
```bash
git commit -m "feat: add dark mode toggle"
git commit -m "fix: resolve login redirect issue"
git commit -m "docs: update API documentation"
```

### Manual Security Scanning

```bash
# Vulnerability scanning with Trivy
trivy fs --severity HIGH,CRITICAL .

# Full Go security report
gosec -fmt=html -out=gosec-report.html ./...
```

## API Endpoints

### Authentication
- `POST /api/auth/google` - Google OAuth login
- `GET /api/auth/me` - Get current user

### Family
- `GET /api/families` - List user's families
- `POST /api/families` - Create family
- `POST /api/families/:id/children` - Add child
- `PUT /api/families/:id/children/:childId` - Update child

### Feeding
- `GET /api/feedings` - List feedings
- `POST /api/feedings` - Create feeding
- `PUT /api/feedings/:id` - Update feeding
- `DELETE /api/feedings/:id` - Delete feeding

### Sleep
- `GET /api/sleep` - List sleep records
- `POST /api/sleep` - Start sleep session
- `PUT /api/sleep/:id` - Update/end sleep
- `DELETE /api/sleep/:id` - Delete sleep record

### Medications
- `GET /api/medications` - List medications
- `POST /api/medications` - Create medication
- `PUT /api/medications/:id` - Update medication
- `DELETE /api/medications/:id` - Delete medication
- `POST /api/medications/:id/deactivate` - Deactivate medication
- `POST /api/medications/log` - Log a dose
- `GET /api/medications/:id/logs` - Get dose history

### Vaccinations
- `GET /api/vaccinations` - List vaccinations
- `POST /api/vaccinations` - Create vaccination
- `PUT /api/vaccinations/:id` - Update vaccination
- `DELETE /api/vaccinations/:id` - Delete vaccination
- `POST /api/vaccinations/generate` - Generate CDC schedule

### Appointments
- `GET /api/appointments` - List appointments
- `POST /api/appointments` - Create appointment
- `PUT /api/appointments/:id` - Update appointment
- `DELETE /api/appointments/:id` - Delete appointment

### Notes
- `GET /api/notes` - List notes
- `POST /api/notes` - Create note
- `PUT /api/notes/:id` - Update note
- `DELETE /api/notes/:id` - Delete note

### Sync
- `POST /api/sync` - Sync offline changes

## Configuration

Configuration is managed via YAML files in `configs/`:

```yaml
server:
  port: 8080
  base_url: http://localhost:8080

database:
  dsn: postgres://babytrack:babytrack123@localhost:5432/babytrack?sslmode=disable

auth:
  google_client_id: your-google-client-id
  google_client_secret: your-google-client-secret
  jwt_secret: your-jwt-secret-change-this-in-production

notifications:
  enabled: false
```

## Roadmap

- [ ] Email invites - Send family invite links via email

## License

MIT
