# MangaHub CLI - Testing Guide

## Quick Start

### Prerequisites

**Terminal 1 - Start the API Server:**

```powershell
cd c:\STUDY\Net_centric\mangahub-v3
go run ./cmd/api-server
# Server will be ready on http://localhost:8080
```

**Terminal 2 - Run CLI Commands:**
All commands below should be run in a separate terminal while the API server is running.

### Build the CLI (Optional)

```powershell
go build -o mangahub.exe ./cmd/cli
```
# Fetch 100 additional series (popular) and just print them
go run ./cmd/cli manga dex

# Fetch 100 series and store them for seeding / integration
go run ./cmd/cli manga dex --output data/manga_api.json

# Search MangaDex by title and show 20 results
go run ./cmd/cli manga dex "attack on titan" --limit 20
### Run Commands Directly

```powershell
go run ./cmd/cli --help
go run ./cmd/cli --version
```

## Command Testing

### Authentication

**Step 1: Register a new user**

```powershell
go run ./cmd/cli auth register --username john3 --email john3@example.com
# When prompted:
# Password: (enter password)
# Confirm Password: (re-enter same password)
```

**Step 2: Login**

````powershell
go run ./cmd/cli --profile user1 auth login --username john2
# When prompted:
# Password: Thienphu123
go run ./cmd/cli --profile user2 auth login --username john3
# Password: Thienphu123
go run ./cmd/cli --profile user3 auth login --username john3

**Step 3: Check authentication status**
```powershell
# If you logged in with --profile, always use the same profile for status and all authenticated commands:
go run ./cmd/cli --profile user2 auth status - john3
go run ./cmd/cli --profile user1 auth status - john2

# Shows current user info and token expiration
````

**Step 4: Logout**

```powershell
go run ./cmd/cli auth logout
# Clears the local session
```

### Manga Search

```powershell
# Basic search
go run ./cmd/cli manga search "attack on titan"

# Search with filters
go run ./cmd/cli manga search "action" --genre action --status ongoing --limit 5

# View details
go run ./cmd/cli manga info attack-on-titan

# List manga
go run ./cmd/cli manga list --page 1 --limit 20

# List with filters
go run ./cmd/cli manga list --status ongoing --limit 10
go run ./cmd/cli manga list --genre shounen --limit 15

# Advanced search
go run ./cmd/cli manga advanced-search --genre shounen --status completed --limit 10
go run ./cmd/cli manga advanced-search "romance" --min-chapters 50 --sort-by title --order asc
```

### Library Management

```powershell

# Add to library (if you logged in with --profile, include it here too)
go run ./cmd/cli --profile user2 library add --manga-id one-piece --status reading --rating 9
go run ./cmd/cli --profile user2 library add --manga-id naruto


# View library
go run ./cmd/cli --profile user2 library list

# Filter library
go run ./cmd/cli --profile user2 library list --status reading
go run ./cmd/cli --profile user1 library list --status reading
go run ./cmd/cli --profile user2 library list --status completed --sort-by title --order asc


# Update entry
go run ./cmd/cli --profile user2 library update --manga-id one-piece --status completed --rating 10
go run ./cmd/cli --profile user1 library update --manga-id one-piece --status completed --rating 10
# Remove from library
go run ./cmd/cli --profile user2 library remove --manga-id one-piece
```

### Progress Tracking

```powershell
# Update progress
go run ./cmd/cli --profile user2 progress update --manga-id naruto --chapter 700
# With additional info
go run ./cmd/cli --profile user2 progress update --manga-id one-piece --chapter 1095 --notes "Epic chapter!"

# View progress history
go run ./cmd/cli --profile user2 progress history

# Manual sync with server
go run ./cmd/cli --profile user2 progress sync


```

### TCP Synchronization (TCP Sync)

```powershell
# Make sure TCP server is running (separate terminal)
go run ./cmd/tcp-server

# Connect to TCP sync server
go run ./cmd/cli sync connect

# Check TCP sync status
go run ./cmd/cli sync status

# Monitor real-time progress updates (press Ctrl+C to stop)
go run ./cmd/cli sync monitor

# Simulate explicit disconnect
go run ./cmd/cli sync disconnect
```

### Notifications

```powershell
# Subscribe to chapter release notifications
go run ./cmd/cli notify subscribe

# Subscribe and keep listening for real-time notifications
go run ./cmd/cli notify subscribe --listen

# Subscribe to specific manga notifications
go run ./cmd/cli notify subscribe --manga-id one-piece

# Unsubscribe from notifications
go run ./cmd/cli notify unsubscribe

# Unsubscribe from specific manga
go run ./cmd/cli notify unsubscribe --manga-id one-piece

# View notification preferences
go run ./cmd/cli notify preferences

# Test notification system
go run ./cmd/cli notify test
```

UDP server-client function (quick path):

- The UDP server runs on port 9091; clients send `register` to subscribe for broadcasts.
- Any JSON payload sent to the server is rebroadcast to all registered clients.
- The CLI `notify subscribe|unsubscribe|preferences|test` flow exercises the same UDP channel.

UDP notifications (manual test):

```powershell
# Start UDP server (separate terminal)
go run ./cmd/udp-server

# Start a UDP listener client (PowerShell)
$udp = New-Object System.Net.Sockets.UdpClient(0)
$udp.Connect('127.0.0.1',9091)
$udp.Send([Text.Encoding]::UTF8.GetBytes('register'),8)  # register client
while($true){$ep=$null; $bytes=$udp.Receive([ref]$ep); [Text.Encoding]::UTF8.GetString($bytes)}

# Fire a notification (PowerShell)
$payload = @{ type='chapter_release'; manga_id='one-piece'; message='Chapter 1101 released'; timestamp=[int][double]::Parse((Get-Date -UFormat %s)) } | ConvertTo-Json
$client = New-Object System.Net.Sockets.UdpClient
$client.Send([Text.Encoding]::UTF8.GetBytes($payload), $payload.Length, '127.0.0.1', 9091) | Out-Null
```

### gRPC Service Operations

```powershell
# Query manga via gRPC
go run ./cmd/cli grpc manga get --id one-piece

# Search via gRPC
go run ./cmd/cli grpc manga search --query "naruto"

# Update progress via gRPC
go run ./cmd/cli grpc progress update --manga-id one-piece --chapter 1095
```

### Chat System

```powershell
# Join general chat
go run ./cmd/cli --profile user1 chat join
go run ./cmd/cli --profile user2 chat join
go run ./cmd/cli --profile user3 chat join
# Join specific manga discussion
go run ./cmd/cli --profile user1 chat join --manga-id aot
go run ./cmd/cli --profile user2 chat join --manga-id aot

# Example
go run ./cmd/cli chat join --manga-id one-piece

# Send message to current chat
go run ./cmd/cli --profile user1 chat send "Hello everyone!"


# Send message to specific manga chat
go run ./cmd/cli --profile user1 chat send "Great chapter!" --manga-id one-piece
go run ./cmd/cli --profile user2 chat send "Great chapter!" --manga-id aot

# View recent messages
go run ./cmd/cli --profile user1 chat history

# View messages for specific manga
go run ./cmd/cli --profile user2 chat history --manga-id aot --limit 50
```

Chat commands (interactive mode):

```
/help  - Show chat commands
/users - List online users
/quit  - Leave chat
/pm <username> <message> - Private message
/manga <id> - Switch to manga-specific chat
/history - Show recent history
/status  - Connection status
```

Expected output (example) for `mangahub chat join`:

```
Connecting to WebSocket chat server at ws://localhost:9093...
✓ Connected to General Chat
Chat Room: #general
Connected users: 12
Your status: Online
Recent messages:
[16:45] alice: Just finished reading the latest chapter!
[16:47] bob: Which manga are you reading?
[16:48] alice: Attack on Titan, it's getting intense
[16:50] charlie: No spoilers please!
─────────────────────────────────────────────────────────────
You are now in chat. Type your message and press Enter.
Type /help for commands or /quit to leave.
johndoe>

johndoe> /help
Chat Commands:
 /help - Show this help
 /users - List online users
 /quit - Leave chat
 /pm <user> <msg>- Private message
 /manga <id> - Switch to manga chat
 /history - Show recent history
 /status - Connection status

johndoe> /users
Online Users (12):
● alice (General Chat)
● bob (General Chat)
● charlie (General Chat)
● diana (One Piece Discussion)
● elena (Attack on Titan Discussion)
● frank (General Chat)
[... 6 more users]

johndoe> Hello everyone!
[17:02] johndoe: Hello everyone!
[17:02] alice: Hey johndoe! Welcome to the chat
[17:03] bob: Hi there! What are you reading these days?

johndoe> /quit
Leaving chat...
✓ Disconnected from chat server
```

### Statistics

```powershell
# Overview
go run ./cmd/cli --profile user2 stats overview

go run ./cmd/cli --profile user1 stats overview

# Detailed
go run ./cmd/cli --profile user2 stats detailed
```

### Data Export

```powershell
# Export library
go run ./cmd/cli  --profile user2  export library --format json --output my-library.json

# Export progress
go run ./cmd/cli  --profile user2  export progress --format csv --output progress.csv

# Full backup
go run ./cmd/cli  --profile user2  export all --output mangahub-backup.tar.gz
```

### Configuration

```powershell
# View config
go run ./cmd/cli config show

# Set value
go run ./cmd/cli config set api-url http://localhost:8080

# Reset
go run ./cmd/cli config reset
```

### Server Management

```powershell
# Check server health
go run ./cmd/cli server health

# View server logs
go run ./cmd/cli server logs
```
go run ./cmd/cli server status

go run ./cmd/cli server start

go run ./cmd/cli server stop


### Database Management

```powershell
# Check database
go run ./cmd/cli db check

# Database stats
go run ./cmd/cli db stats

# Repair database
go run ./cmd/cli db repair
```
go run ./cmd/cli db optimize
## Testing Output Examples

### Successful Search

```
Searching for "attack on titan"...

Found 1 results:

┌──────────────────────────────────────────────────────────────────────────────────────────┐
│ ID   │ TITLE                          │ AUTHOR               │ STATUS     │ CHAPTERS │
├──────────────────────────────────────────────────────────────────────────────────────────┤
│ a... │ Attack on Titan                │ Hajime Isayama       │ completed  │      141 │
└──────────────────────────────────────────────────────────────────────────────────────────┘

Use 'mangahub manga info <id>' to view details
Use 'mangahub library add --manga-id <id>' to add to your library
```

### Manga List

```
Listing manga (page 1, limit 15)

┌────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ ID     │ TITLE                               │ AUTHOR                    │ STATUS     │ CHAPTERS │
├────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ 20t... │ 20th Century Boys                   │ Naoki Urasawa             │ completed  │      249 │
│ one... │ One Piece                           │ Eiichiro Oda              │ ongoing    │     1100 │
│ nar... │ Naruto                              │ Masashi Kishimoto         │ completed  │      700 │
└────────────────────────────────────────────────────────────────────────────────────────────────────────┘

Showing 15 manga (page 1)
Use --page <n> to see more results
```

### Manga Info

```
┌──────────────────────────────────────────────────────────────────────┐
│ ONE PIECE                                                            │
└──────────────────────────────────────────────────────────────────────┘

Basic Information:
  ID:      one-piece
  Title:   One Piece
  Author:  Eiichiro Oda
  Genres:  Action, Adventure, Comedy, Shounen
  Status:  ongoing

Progress:
  Total Chapters: 1100
  Created:        2025-12-17
  Updated:        2025-12-17

Description:
  Monkey D. Luffy sets off on an adventure to find the legendary One Piece
  treasure and become the Pirate King.

Actions:
  Update Progress: mangahub progress update --manga-id one-piece --chapter <n>
  Add to Library:  mangahub library add --manga-id one-piece
  Rate/Review:     mangahub library update --manga-id one-piece --rating <1-10>
```

### Auth Status

```
Authentication Status: ✓ Logged in

User ID:  user_abc123
Username: johndoe
Email:    john@example.com
Expires:  2025-12-19 09:30:00 UTC
```

## Global Flags Testing

### All commands support:

```powershell
# With auth token
go run ./cmd/cli manga search "naruto" --token your-token-here

# With custom API URL
go run ./cmd/cli manga search "naruto" --api http://localhost:8080

# With verbose output
go run ./cmd/cli manga search "naruto" --verbose

# Show help
go run ./cmd/cli manga search --help
```

## Error Testing

### Invalid flags

```powershell
go run ./cmd/cli library add                    # Error: --manga-id required
go run ./cmd/cli progress update                # Error: --manga-id and --chapter required
go run ./cmd/cli auth login                     # Error: --username or --email required
```

### Command help

```powershell
go run ./cmd/cli --help                         # Show all commands
go run ./cmd/cli manga --help                   # Show manga subcommands
go run ./cmd/cli manga search --help            # Show search flags
go run ./cmd/cli auth --help                    # Show auth subcommands
```

## Integration Testing Checklist

### Core Infrastructure

- [x] All commands execute without syntax errors
- [x] Help text displays correctly
- [x] Flags are parsed correctly
- [x] Global flags work on all commands
- [x] Error messages are user-friendly
- [x] Output formatting is readable

### Manga Operations (No Auth Required)

- [x] Manga search queries API server
- [x] Manga list with pagination works
- [x] Manga info displays details
- [x] Advanced search with filters works

### Authentication (API Server Required)

- [x] Auth register creates users via HTTP API
- [x] Auth login validates credentials via HTTP API
- [x] Auth status validates token with API server
- [x] Auth logout clears local session
- [x] Session saved to ~/.mangahub/session.json

### Authenticated Features

- [x] Library management (add, list, update, remove)
- [x] Progress tracking (update, history)
- [ ] Notifications (subscribe, preferences, test, unsubscribe)
- [ ] Chat functionality (needs WebSocket server)
- [ ] gRPC operations (needs gRPC server)

## Running Servers

```powershell
# API Server (HTTP REST) - Port 8080
go run ./cmd/api-server

# gRPC Server - Port 9092
go run ./cmd/grpc-server

# TCP Server - Port 9090
go run ./cmd/tcp-server

# UDP Server - Port 9091
go run ./cmd/udp-server

# WebSocket Server - Port 9093
go run ./cmd/websocket-server
```

## Quick Test Commands

```powershell
# Test manga functionality
go run ./cmd/cli manga list --limit 10
go run ./cmd/cli manga search "one piece"
go run ./cmd/cli manga info one-piece

# Test auth functionality
go run ./cmd/cli auth status
go run ./cmd/cli auth register --username testuser --email test@example.com
go run ./cmd/cli auth login --username testuser

# Test advanced search
go run ./cmd/cli manga advanced-search --genre shounen --status completed --limit 5
```

---

**Database**: 200 manga entries (100 manual + 100 API)
**Status**: CLI connected to SQLite database
**Config**: config.yaml for server settings
