# ADR-0005: Adopt Standard Go Project Structure

- Status: accepted
- Date: 2026-01-08
- Deciders: Sergey Grachev

## Context

The KBase Image Catalog project requires a well-organized project structure to ensure maintainability, scalability, and developer productivity. The choice of directory layout impacts code organization, build process, testing strategy, and overall project governance.

Key considerations:
- **Go Conventions**: Following Go community standards and best practices
- **Maintainability**: Clear separation of concerns and easy navigation
- **Scalability**: Structure that supports growth and feature additions
- **Testing**: Clear organization for unit and integration tests
- **Build Process**: Simple and predictable build artifacts
- **Documentation**: Easy location and maintenance of project docs
- **Team Collaboration**: Consistent structure for multiple developers
- **Tooling Support**: Compatibility with Go tooling and IDEs
- **Dependency Management**: Clear module boundaries and interfaces
- **Deployment**: Simple packaging and distribution

## Decision

We will adopt the **standard Go project structure** following the conventions established by the Go community and documented in "How to Write Go Code" and "Project Layout" recommendations.

## Consequences

### Positive Consequences

1. **Community Standards**: Follows established Go conventions and best practices
2. **Developer Familiarity**: Easy for Go developers to understand and navigate
3. **Tooling Compatibility**: Works seamlessly with Go toolchain (go build, go test, etc.)
4. **Clear Separation**: Well-defined boundaries between different concerns
5. **Simplicity**: Minimal complexity with straightforward organization
6. **Scalability**: Structure naturally supports project growth
7. **Testing Organization**: Clear separation of test code from production code
8. **Documentation Integration**: Standard location for project documentation
9. **IDE Support**: Excellent integration with Go IDEs (VS Code, GoLand, etc.)
10. **Build Predictability**: Consistent and predictable build artifacts

### Negative Consequences

1. **Initial Learning Curve**: Developers new to Go need to understand the conventions
2. **Less Flexible**: More rigid structure compared to some frameworks
3. **Hidden Files**: `internal` directory hides implementation details from external consumers
4. **Template Location**: Templates are separated from application code
5. **Build Directory**: Separate build artifacts from source code
6. **Package Organization**: Requires careful consideration of package boundaries
7. **Dependency Management**: `go.mod` and `go.sum` files at project root
8. **Configuration Files**: Configuration files mixed with source code
9. **Static Assets**: Web assets separated from application code
10. **Limited Customization**: Less room for custom organizational patterns

## Alternatives Considered

### 1. Monolithic Structure (All in `src/`)
**Rejected** because:
- Violates Go conventions
- Makes dependency management difficult
- Poor separation of concerns
- Harder to test and maintain
- Not idiomatic Go practice

### 2. Domain-Driven Design (DDD) Structure
**Rejected** because:
- Over-engineered for current project size
- Complex package boundaries
- Steeper learning curve for team
- More ceremony than needed for catalog application
- Potential over-abstraction

### 3. Feature-Based Structure
**Rejected** because:
- Can lead to circular dependencies
- Harder to find related functionality
- Poor separation of technical concerns
- Makes testing more complex
- Not aligned with Go idioms

### 4. Service-Oriented Structure
**Rejected** because:
- Premature for current project scope
- Creates artificial boundaries
- More complex dependency management
- Overkill for monolithic application
- Harder to reason about code organization

### 5. Framework-Based Structure
**Rejected** because:
- Limits flexibility and customization
- Creates framework dependency
- May not align with Go conventions
- Harder to migrate away from framework
- Adds unnecessary complexity

## Related ADRs

- [ADR-0002: Use Go for Backend Development](./ADR-0002-use-go-for-backend.md) - Backend technology that uses this structure
- [ADR-0003: Use HTMX for Web Interface](./ADR-0003-htmx-for-web-interface.md) - Web interface assets organization
- [ADR-0004: Integrate LLM Models for Image Analysis](./ADR-0004-llm-integration.md) - Analysis components organization
- [ADR-0006: Implement Comprehensive Error Handling Strategy](./ADR-0006-error-handling.md) - Error handling patterns

## Implementation Details

### Standard Go Project Structure

```
kbase-catalog/
├── cmd/kbase-catalog/          # Application entry point
│   └── main.go                 # Main application file
├── internal/                   # Private application code
│   ├── config/                 # Configuration management
│   │   └── config.go           # Configuration loading and validation
│   ├── encoder/                # Image encoding utilities
│   │   └── encoder.go          # Image processing functions
│   ├── llm/                    # LLM integration
│   │   └── llm.go              # AI model client
│   ├── processor/              # Core processing logic
│   │   ├── processor.go        # Main processor coordination
│   │   ├── directory_processor.go # Directory processing
│   │   ├── file_scanner.go     # File system scanning
│   │   ├── image_processor.go  # Image analysis
│   │   └── index_generator.go  # Index file generation
│   └── web/                    # Web interface components
│       ├── server.go           # HTTP server
│       ├── services/           # Business logic
│       │   ├── catalog_service.go
│       │   └── template_renderer.go
│       ├── queue/              # Task processing
│       │   └── queue.go
│       ├── watch/              # File monitoring
│       │   └── watch.go
│       ├── middleware.go
│       └── sorting.go
├── web/                        # Web assets
│   ├── static/                 # CSS, JS, images
│   │   ├── styles.css
│   │   ├── htmx.min.js
│   │   ├── viewer.min.js
│   │   └── viewer.min.css
│   └── templates/              # HTML templates
│       ├── index.html
│       ├── catalog-detail.html
│       ├── search-result.html
│       └── *-fragment.html
├── api/                        # API definitions (if needed)
│   └── openapi.yaml
├── configs/                    # Configuration files
│   ├── default.yaml
│   ├── production.yaml
│   └── development.yaml
├── deployments/                # Deployment configurations
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   ├── kubernetes/
│   │   └── k8s.yaml
│   └── systemd/
│       └── kbase-catalog.service
├── docs/                       # Documentation
│   ├── adr/                    # Architecture Decision Records
│   ├── api/                    # API documentation
│   ├── user-guide/             # User documentation
│   └── development/            # Developer documentation
├── examples/                   # Example code and usage
│   ├── basic-usage/
│   └── advanced-configuration/
├── test/                       # Additional test utilities
│   ├── integration/
│   ├── fixtures/
│   └── mockdata/
├── tools/                      # Development tools
│   ├── build.sh
│   ├── test.sh
│   └── lint.sh
├── scripts/                    # Utility scripts
│   ├── migrate.sh
│   ├── backup.sh
│   └── deploy.sh
├── pkg/                        # Public packages (if needed)
│   └── utils/                  # Utility functions
├── .gitignore                  # Git ignore rules
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
├── Makefile                    # Build automation
├── README.md                   # Project documentation
├── LICENSE                     # License file
├── config.yaml                 # Default configuration
└── build-linux.sh              # Build script
```

### Package Organization Principles

1. **Single Responsibility**: Each package has a clear, focused purpose
2. **Loose Coupling**: Minimal dependencies between packages
3. **High Cohesion**: Related functionality grouped together
4. **Interface Segregation**: Clear interfaces for external dependencies
5. **Dependency Inversion**: Depend on abstractions, not concretions

### Naming Conventions

1. **Package Names**: Lowercase, descriptive, singular nouns
2. **File Names**: Lowercase, matching package name when appropriate
3. **Interface Names**: `-able`, `-er` suffixes for behavior contracts
4. **Error Types**: Descriptive names ending in "Error"
5. **Configuration**: YAML files with environment-specific names

### Build and Deployment

1. **Binary Output**: Single binary in project root
2. **Cross-Platform**: Build scripts for multiple platforms
3. **Container Support**: Dockerfile for containerized deployment
4. **Package Management**: Standard Go modules
5. **Static Linking**: Option for static binary builds

## Testing Organization

```
internal/
├── config/
│   ├── config.go
│   └── config_test.go
├── processor/
│   ├── processor.go
│   ├── processor_test.go
│   └── processor_integration_test.go
test/
├── integration/
│   ├── catalog_processing_test.go
│   └── web_interface_test.go
├── fixtures/
│   ├── sample_images/
│   └── sample_config.yaml
└── mockdata/
    └── mock_responses.json
```

## Future Considerations

- **Microservices Migration**: Structure supports future service decomposition
- **Plugin Architecture**: Extensible for future plugin system
- **Multi-Module Growth**: Structure supports multiple Go modules
- **Documentation Growth**: Scalable documentation organization
- **Testing Expansion**: Room for comprehensive test coverage

## Maintenance Guidelines

1. **Regular Reviews**: Periodically review package boundaries
2. **Refactoring**: Continuously improve code organization
3. **Documentation**: Keep documentation updated with structure changes
4. **Tooling**: Use Go tooling to maintain code quality
5. **Team Input**: Solicit feedback from all team members on structure effectiveness

## Migration Considerations

- **Legacy Code**: Migrate existing code to new structure incrementally
- **Build Scripts**: Update build scripts to match new structure
- **Documentation**: Update all references to old structure
- **CI/CD**: Update deployment scripts for new layout
- **Team Training**: Ensure all team members understand new structure