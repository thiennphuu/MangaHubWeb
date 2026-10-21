# MangaHub

A comprehensive manga library management system built with Go, featuring multi-protocol communication (HTTP REST API, TCP, UDP, gRPC, WebSocket), CLI interface, and real-time synchronization capabilities.

## Features

### Core Functionality

- **User Management**: Multi-profile support, JWT-based authentication, secure password hashing
- **Manga Library Management**: Add, remove, update, and track manga in personal libraries
- **Reading Progress Tracking**: Track reading progress across devices with real-time synchronization
- **Statistics & Analytics**: Overview and detailed statistics for reading habits
- **Data Export/Import**: Export library and progress data to JSON/CSV formats

### Network Protocols

- **HTTP REST API** (Port 8080): 20+ REST endpoints for all operations
- **TCP Sync Server** (Port 9090): Real-time progress synchronization across devices
- **UDP Notification Server** (Port 9091): Lightweight, connectionless notifications
- **gRPC Service** (Port 9092): High-performance RPC with Protocol Buffers
- **WebSocket Chat** (Port 9093): Real-time room-based chat system

### Advanced Features

- **MangaDex API Integration**: Fetch manga metadata from MangaDex API
- **Server Management**: Database operations (check, optimize, repair, stats), log viewing
- **Community Chat**: Room-based real-time messaging via WebSocket
- **Notification System**: Subscribe to manga updates with UDP notifications
- **Backup & Restore**: Create and restore system backups

## Architecture

### Server Components

- **API Server** (`cmd/api-server`): HTTP REST API server with Gin framework
- **TCP Server** (`cmd/tcp-server`): TCP-based synchronization server
- **UDP Server** (`cmd/udp-server`): UDP notification broadcast server
- **gRPC Server** (`cmd/grpc-server`): gRPC service for manga operations
- **WebSocket Server** (`cmd/websocket-server`): WebSocket chat server

### Client

- **CLI Application** (`cmd/cli`): Comprehensive command-line interface with 50+ commands

### Data Storage

- **SQLite Database**: Local SQLite database for all data persistence
- **Session Storage**: JSON-based session files for multi-profile support

## Installation

### Prerequisites

- Go 1.24 or higher
- CGO enabled (for SQLite support)
- Git

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd MangaHub

# Install dependencies
go mod download

# Build all servers
go build -o bin/api-server ./cmd/api-server
go build -o bin/tcp-server ./cmd/tcp-server
go build -o bin/udp-server ./cmd/udp-server
go build -o bin/grpc-server ./cmd/grpc-server
go build -o bin/websocket-server ./cmd/websocket-server

# Build CLI
go build -o bin/mangahub ./cmd/cli
```

### Docker Deployment

```bash
# Build Docker image
docker build -t mangahub .

# Run with docker-compose
docker-compose up -d
```

## Quick Start

### 1. Configure the Server

Edit `config.yaml` to set your server IP and ports:

```yaml
http:
  host: 10.238.53.72 # Your server IP
  port: 8080

tcp:
  host: 10.238.53.72
  port: 9090
# ... other protocols
```

### 2. Start the Servers

```bash
# Start API server
./bin/api-server

# Start TCP server
./bin/tcp-server

# Start UDP server
./bin/udp-server

# Start gRPC server
./bin/grpc-server

# Start WebSocket server
./bin/websocket-server
```

### 3. Use the CLI

```bash
# Register a new user
./bin/mangahub auth register --username alice --password secret123

# Login
./bin/mangahub auth login --username alice --password secret123

# List available manga
./bin/mangahub manga list

# Add manga to library
./bin/mangahub library add --manga-id one-piece

# Update reading progress
./bin/mangahub progress update --manga-id one-piece --chapter 1000

# View statistics
./bin/mangahub stats overview
```

## Configuration

Configuration is managed via `config.yaml`:

```yaml
app:
  name: MangaHub
  version: 1.0.0
  environment: development
  jwt_secret: your-secret-key-change-in-production

database:
  type: sqlite
  path: data/mangahub.db
  auto_migrate: true

http:
  host: 10.238.53.72
  port: 8080

tcp:
  host: 10.238.53.72
  port: 9090
  max_connections: 50

udp:
  host: 10.238.53.72
  port: 9091
  max_clients: 100

grpc:
  host: 10.238.53.72
  port: 9092

websocket:
  host: 10.238.53.72
  port: 9093
```

Environment variables can override configuration values (e.g., `MANGAHUB_API_URL`, `TCP_SERVER_HOST`).

## Project Structure

```
MangaHub/
├── cmd/                    # Application entry points
│   ├── api-server/        # HTTP REST API server
│   ├── cli/               # CLI application
│   ├── tcp-server/        # TCP sync server
│   ├── udp-server/        # UDP notification server
│   ├── grpc-server/       # gRPC service
│   └── websocket-server/  # WebSocket chat server
│
├── internal/              # Internal application code
│   ├── api/              # HTTP API handlers
│   ├── auth/             # Authentication logic
│   ├── cli/              # CLI command implementations
│   ├── grpc/             # gRPC service implementations
│   ├── manga/            # Manga business logic
│   ├── tcp/              # TCP server implementation
│   ├── udp/              # UDP server implementation
│   ├── user/             # User service logic
│   └── websocket/        # WebSocket hub and handlers
│
├── pkg/                   # Public packages
│   ├── client/           # Client libraries (HTTP, TCP, UDP, gRPC, WebSocket)
│   ├── config/           # Configuration management
│   ├── database/         # Database layer (SQLite)
│   ├── models/           # Data models
│   ├── output/           # Output formatters
│   ├── session/          # Session management
│   └── utils/            # Utility functions
│
├── proto/                 # Protocol Buffer definitions
│   └── manga.proto       # gRPC service definitions
│
├── data/                  # Database and data files
├── logs/                  # Server logs
├── scripts/               # Helper scripts
├── config.yaml           # Configuration file
└── README.md             # This file
```

## CLI Commands Overview

### Authentication

- `mangahub auth register` - Register a new user
- `mangahub auth login` - Login to the system
- `mangahub auth logout` - Logout from current session
- `mangahub auth status` - Check authentication status
- `mangahub auth change-password` - Change user password

### Profile Management

- `mangahub profile create` - Create a new profile
- `mangahub profile list` - List all profiles
- `mangahub profile switch` - Switch active profile

### Manga Operations

- `mangahub manga list` - List all available manga
- `mangahub manga info` - Get detailed manga information
- `mangahub manga search` - Search manga by title
- `mangahub manga advanced-search` - Advanced search with filters
- `mangahub manga dex` - Fetch manga from MangaDex API

### Library Management

- `mangahub library list` - List manga in your library
- `mangahub library add` - Add manga to library
- `mangahub library remove` - Remove manga from library
- `mangahub library update` - Update library entry

### Progress Tracking

- `mangahub progress update` - Update reading progress
- `mangahub progress history` - View progress history
- `mangahub progress status` - Get current progress status
- `mangahub progress sync` - Sync progress via HTTP API

### TCP Synchronization

- `mangahub sync connect` - Connect to TCP sync server
- `mangahub sync status` - Check sync connection status
- `mangahub sync monitor` - Monitor real-time sync updates
- `mangahub sync disconnect` - Disconnect from sync server

### UDP Notifications

- `mangahub notify subscribe` - Subscribe to manga notifications
- `mangahub notify unsubscribe` - Unsubscribe from notifications
- `mangahub notify preferences` - Manage notification preferences
- `mangahub notify test` - Test notification system

### WebSocket Chat

- `mangahub chat join` - Join a chat room
- `mangahub chat send` - Send a message to room
- `mangahub chat history` - View chat history

### gRPC Operations

- `mangahub grpc manga get` - Get manga via gRPC
- `mangahub grpc manga search` - Search manga via gRPC
- `mangahub grpc progress update` - Update progress via gRPC

### Statistics

- `mangahub stats overview` - View reading statistics overview
- `mangahub stats detailed` - View detailed statistics

### Data Export

- `mangahub export library` - Export library to JSON/CSV
- `mangahub export progress` - Export progress to JSON/CSV
- `mangahub export all` - Export all data

### Server Management

- `mangahub server status` - Check server status
- `mangahub server health` - Check server health
- `mangahub server logs` - View server logs
- `mangahub db check` - Check database integrity
- `mangahub db optimize` - Optimize database
- `mangahub db stats` - View database statistics
- `mangahub db repair` - Repair database

### Configuration

- `mangahub config show` - Show current configuration
- `mangahub config set` - Set configuration value
- `mangahub config reset` - Reset configuration

## API Endpoints

### Authentication

- `POST /auth/register` - Register new user
- `POST /auth/login` - User login
- `GET /auth/status` - Check authentication status

### Manga

- `GET /manga` - List all manga
- `GET /manga/:id` - Get manga by ID
- `GET /manga/search` - Search manga

### User

- `GET /users/profile` - Get user profile
- `PUT /users/profile` - Update user profile
- `GET /users/library` - Get user library
- `POST /users/library` - Add manga to library
- `DELETE /users/library/:id` - Remove manga from library
- `PUT /users/library/:id/progress` - Update reading progress

### Server

- `GET /health` - Health check
- `GET /server/logs` - Get server logs
- `GET /server/database/check` - Check database
- `POST /server/database/optimize` - Optimize database
- `GET /server/database/stats` - Database statistics
- `POST /server/database/repair` - Repair database

## Technologies

### Core

- **Go 1.24**: Programming language
- **SQLite**: Database (via `github.com/mattn/go-sqlite3` and `modernc.org/sqlite`)

### Frameworks & Libraries

- **Gin** (`github.com/gin-gonic/gin`): HTTP web framework
- **Cobra** (`github.com/spf13/cobra`): CLI framework
- **gRPC** (`google.golang.org/grpc`): gRPC framework
- **Protocol Buffers** (`google.golang.org/protobuf`): Data serialization
- **gorilla/websocket** (`github.com/gorilla/websocket`): WebSocket implementation
- **JWT** (`github.com/golang-jwt/jwt/v4`): JSON Web Tokens
- **bcrypt** (`golang.org/x/crypto`): Password hashing

### Protocols & Standards

- HTTP/1.1 (REST API)
- HTTP/2 (gRPC)
- TCP (Line-delimited JSON)
- UDP (Datagram-based)
- WebSocket (RFC 6455)
- Protocol Buffers v3
- JSON
- JWT (RFC 7519)

## Development

### Generate gRPC Code

```bash
# Install protoc compiler and Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code from proto files
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/manga.proto
```

### Running Tests

```bash
go test ./...
```

### Code Style

The project follows standard Go code style. Use `gofmt` and `golint`:

```bash
gofmt -w .
golint ./...
```

## License

MIT License

Copyright (c) 2024 MangaHub Contributors
