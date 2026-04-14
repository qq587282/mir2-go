# AGENTS.md - Development Guidelines for Mir2-Go
使用 Go 语言重写的热血传奇2/3 (Legend of Mir 2/3) 游戏服务器引擎。

This file provides guidelines for AI coding agents operating in this Mir2-Go game server project.

## Project Overview

Mir2-Go is a Go-based reimplementation of the Legend of Mir 2/3 game server. The project follows standard Go patterns and conventions.

## Build, Lint, and Test Commands

### Go Commands
```bash
# Build all packages
go build ./...

# Build specific binary
go build -o bin/m2server ./cmd/m2server

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a single test file
go test -v ./pkg/game/actor/player_test.go

# Run tests matching pattern
go test -v -run "TestPlayer" ./pkg/game/actor/

# Run tests with coverage
go test -v -cover ./...

# Run linter (requires golangci-lint)
golangci-lint run继续

# Format code
go fmt ./...

# Vet code
go vet ./...

# Tidy go.mod
go mod tidy
```

### Build Scripts
```bash
# Windows
build.bat

# Linux/Mac
./build.sh
```

## 环境
这是一个移植项目，移植前的代码目录在D:\code\mir2
修改代码时候要完美移植游戏的玩法和功能，保留客户端和服务器的交互逻辑
Go 安装目录在 D:\Program Files\go


## 重要
兼容源代码协议和客户端，保留源代码游戏的玩法逻辑，每次修改有需要更新readme文件

## Code Style Guidelines

### Imports
- Use standard Go import organization (stdlib, external, internal)
- Group imports: standard library first, then external packages, then internal
- Use canonical import paths: `github.com/mir2go/mir2/pkg/...`
- Alphabetize within groups

### Formatting
- Use `gofmt` or `go fmt` for formatting
- 4-space indentation (Go standard, not 2-space)
- Maximum line length: 120 characters
- No trailing commas (Go format handles this)
- Use blank lines to separate logical sections

### Types
- Use explicit types rather than `var` inference
- Prefer interfaces over concrete types when polymorphism is needed
- Use specific types: `int32`, `uint16` for protocol data (matching Mir2 protocol)
- Return explicit types for exported functions

### Naming Conventions
- **Packages**: lowercase, short names (e.g., `actor`, `guild`, `ranking`)
- **Types/Structs**: PascalCase (e.g., `Player`, `TMonster`, `GameMap`)
- **Interfaces**: PascalCase with `er` suffix where appropriate (e.g., `Reader`, `Database`)
- **Functions/Methods**: PascalCase for exported, camelCase for unexported
- **Constants**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase, avoid single letters except in loops
- **Files**: lowercase, snake_case (e.g., `player.go`, `guild_manager.go`)

### Error Handling
- Always handle errors explicitly - never ignore with `_`
- Return errors with context using `fmt.Errorf("failed to x: %w", err)`
- Use custom error types for domain-specific errors
- Never log and return the same error
- Use `defer` for cleanup (file handles, mutex unlocking)

### Comments
- **DO NOT ADD COMMENTS** unless explicitly requested by the user
- Write self-documenting code with descriptive names
- Add godoc comments only for exported APIs that need documentation
- No inline comments explaining obvious code

### General Principles
- Keep functions small (< 50 lines when possible)
- Single responsibility principle
- Favor composition over inheritance
- Use interfaces for dependency injection
- Keep package internal state private
- Use sync.Pool for frequently allocated objects
- Follow Go proverb: "Clear is better than clever"

## Project Structure

```
mir2-go/
├── cmd/           # Application entry points (servers)
├── pkg/          # Core packages
│   ├── protocol/ # Network protocol definitions
│   ├── network/  # Networking layer
│   ├── config/   # Configuration
│   ├── game/     # Game logic
│   │   ├── actor/    # Player, monster, NPC
│   │   ├── map/      # Map and pathfinding
│   │   ├── skill/    # Skills and magic
│   │   ├── guild/    # Guild system
│   │   ├── quest/    # Quest system
│   │   ├── mail/     # Mail system
│   │   ├── pet/      # Pet system
│   │   └── ranking/  # Rankings
│   ├── db/       # Database layer
│   └── script/   # NPC script engine
├── config.yaml   # Configuration file
├── go.mod        # Go module file
└── README.md     # Project documentation
```

## Key Conventions

### Protocol Compatibility
- This project implements Mir2 network protocol
- Use exact types from `pkg/protocol` for message structures
- Message IDs are defined in `pkg/protocol/message_id.go`
- Binary serialization uses `encoding/binary` with LittleEndian

### Database Patterns
- Database interface defined in `pkg/db/db.go`
- Implementations for MySQL and SQLite
- Use transactions for multi-table operations

### Script Engine
- NPC scripts use custom DSL in `pkg/script/engine.go`
- Follow existing command patterns for new features

### Network Layer
- GateServer handles TCP connections in `pkg/network/gate.go`
- Session management with concurrent-safe maps
- Use channels for async message processing

## Testing Guidelines
- Test files named `*_test.go` in same package
- Use descriptive test names: `TestName_Given_When_Then`
- Follow AAA pattern: Arrange, Act, Assert
- Mock external dependencies
- Create test utilities in `pkg/utils/` if shared

## Git Workflow
- Create feature branches from main
- Commit messages: imperative mood, 50 chars max subject
- Run `go build ./...` and `go test ./...` before committing
- No commit should break the build