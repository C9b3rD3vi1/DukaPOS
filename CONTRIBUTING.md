# DukaPOS - Contributing Guide

Thank you for your interest in contributing to DukaPOS!

## Development Setup

1. **Clone the repository**
```bash
git clone https://github.com/C9b3rD3vi1/DukaPOS.git
cd DukaPOS
```

2. **Install dependencies**
```bash
go mod download
```

3. **Create environment file**
```bash
cp .env.example .env
# Edit .env with your settings
```

4. **Run the application**
```bash
go run cmd/server/main.go
```

5. **Run tests**
```bash
go test ./...
```

## Code Style

- Use meaningful variable names
- Add comments for complex logic
- Follow Go standard conventions
- Run `go fmt` before committing

## Project Structure

```
DukaPOS/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── database/       # Database connection & migrations
│   ├── handlers/      # HTTP handlers
│   ├── middleware/    # Fiber middleware
│   ├── models/        # Data models
│   ├── repository/    # Database operations
│   └── services/     # Business logic
├── static/            # Web dashboard
├── tests/            # Test files
└── migrations/       # SQL migrations
```

## Feature Development

1. Create a feature branch
2. Implement the feature
3. Add tests
4. Update documentation
5. Submit a pull request

## Reporting Issues

Use GitHub Issues to report bugs or request features.

## License

MIT License - see LICENSE file.
