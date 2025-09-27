# Wire Dependency Injection Setup

## Overview

This project uses Google Wire for compile-time dependency injection. Wire generates code that wires up your dependencies automatically, providing type safety and eliminating runtime reflection.

## Architecture

The dependency injection is organized into layers:

### 1. Infrastructure Layer (`internal/infrastructure/providers.go`)
- Database connections (MongoDB, Redis)
- Authentication managers (JWT, Password)
- Validators and i18n services

### 2. Repository Layer (`internal/repository/providers.go`)
- Data access layer implementations
- Provides domain interfaces

### 3. Service Layer (`internal/service/providers.go`)
- Business logic implementations
- Uses repository interfaces

### 4. Handler Layer (`internal/handler/providers.go`)
- HTTP handlers
- Uses service interfaces

### 5. Application Layer (`internal/app/wire.go`)
- Main application wiring
- Combines all provider sets

## Key Files

- `internal/app/wire.go` - Wire injector definitions (build tag: `wireinject`)
- `internal/app/wire_gen.go` - Generated code (build tag: `!wireinject`)
- `internal/*/providers.go` - Provider functions for each layer

## Usage

### Generating Wire Code

```bash
# From backend directory
wire gen ./internal/app

# Or using Makefile from project root
make wire-gen
```

### Adding New Dependencies

1. Create provider function in appropriate layer:
```go
// In internal/service/providers.go
func ProvideNewService(repo domain.Repository) domain.Service {
    return NewService(repo)
}
```

2. Add to provider set:
```go
var ServiceSet = wire.NewSet(
    ProvideUserService,
    ProvidePhotoService,
    ProvideNewService, // Add here
)
```

3. Regenerate Wire code:
```bash
wire gen ./internal/app
```

### Application Initialization

The application is initialized using Wire in `cmd/main.go`:

```go
// Create application with Wire DI
application, err := app.InitializeApp(cfg, logger)
```

## Benefits

1. **Compile-time Safety**: Dependencies are resolved at compile time
2. **No Runtime Reflection**: Better performance than runtime DI
3. **Type Safety**: Compiler catches dependency issues
4. **Clear Dependencies**: Easy to see what each component needs
5. **Testability**: Easy to mock dependencies for testing

## Build Tags

- `//go:build wireinject` - Only compiled when generating Wire code
- `//go:build !wireinject` - Compiled in normal builds (generated code)

## Troubleshooting

### Common Issues

1. **Undefined provider**: Make sure provider function is in correct provider set
2. **Circular dependencies**: Refactor to break cycles
3. **Missing interfaces**: Ensure all dependencies implement required interfaces

### Regenerating Code

Always regenerate Wire code after:
- Adding new providers
- Changing provider signatures
- Modifying dependency structure

```bash
make wire-gen
```

## Testing

Wire makes testing easier by allowing easy dependency injection:

```go
func TestService(t *testing.T) {
    mockRepo := &MockRepository{}
    service := ProvideService(mockRepo)
    // Test service with mock
}
```
