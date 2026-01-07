# KBase Image Catalog - AI-Powered Image Knowledge Management System

<div align="center">

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![GPLv3](https://img.shields.io/badge/License-GPL%20v3-blue?style=for-the-badge)

[//]: # (![Build Status]&#40;https://img.shields.io/badge/Build-Passing-green?style=for-the-badge&#41;)

**🤖 Transform your image collections into intelligent, searchable catalogs using AI/ML**

[📖 Documentation](#-documentation) • [🚀 Quick Start](#-quick-start) • [🔧 Configuration](#-configuration) • [🐳 Docker](#-docker-deployment) • [🤝 Contributing](#-contributing)

</div>

## 📋 Table of Contents

- [📋 Project Overview](#-project-overview)
- [🏗️ Project Architecture](#-project-architecture)
- [🚀 Quick Start](#-quick-start)
    - [Prerequisites](#prerequisites)
    - [Using Docker Compose](#using-docker-compose)
    - [Installation](#installation)
- [📊 Usage Examples](#-usage-examples)
- [🔧 Configuration](#-configuration)
    - [Configuration Parameters](#configuration-parameters)
    - [Advanced Configuration Example](#advanced-configuration-example)
- [🧪 Testing and Development](#-testing-and-development)
- [🐳 Docker Deployment](#-docker-deployment)
- [🚀 Deployment](#-deployment)
- [📈 Monitoring and Logging](#-monitoring-and-logging)
- [🔒 Security](#-security)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)
- [🆘 Support & Community](#-support--community)
- [🎯 Roadmap & Future Development](#-roadmap--future-development)

## 📋 Project Overview

KBase Image Catalog is a production-ready Go application that leverages artificial intelligence to automatically process
and catalog large image collections. The system analyzes each image using advanced LLM models and generates structured
metadata (short names and detailed descriptions), transforming disorganized image libraries into searchable, intelligent
catalogs.

### 🎯 Key Benefits

- **🤖 AI-Powered Analysis**: Automatically generates descriptive metadata using LLaVA and Qwen-VL models
- **⚡ High Performance**: Parallel processing with configurable concurrency for rapid cataloging
- **🌐 Modern Web Interface**: HTMX-powered dynamic interface with real-time search and filtering
- **🔄 Auto-Updates**: File system monitoring for automatic catalog refreshes
- **📊 Comprehensive Metadata**: Structured JSON output with rich information
- **🛡️ Production Ready**: Docker support, graceful shutdown, and comprehensive logging

### 🎯 Key Features

#### 🤖 AI-Powered Processing

- **Image Recognition** using LLM models (LLaVA, Qwen-VL)
- **Metadata Generation** in JSON format with short_name and description
- **Parallel Processing** for high performance
- **Retry Mechanism** with configurable parameters

#### 📁 Smart Directory Management

- **Recursive Scanning** of nested directories
- **Automatic Filtering** of supported image formats
- **index.json Generation** for each catalog
- **Root Index Creation** for the entire collection

#### 🌐 Web Interface

- **Modern UI** using HTMX for dynamic updates
- **Full-Text Search** across entire collection and within individual catalogs
- **Sorting and Filtering** of results
- **Automatic Updates** when files change
- **Image Preview** with integrated viewer

#### ⚡ High Performance

- **Asynchronous Processing** using task queues
- **File System Monitoring** for automatic updates
- **Memory-Optimized** operations
- **Graceful Shutdown** with signal handling

## 🏗️ Project Architecture

### 📁 Project Structure

```
kbase-catalog/
├── cmd/kbase-catalog/          # Application entry point
│   └── main.go                 # Main CLI interface
├── internal/                   # Internal packages (non-exportable)
│   ├── config/                 # Configuration management
│   │   ├── config.go           # Configuration loading and validation
│   │   └── config_test.go      # Configuration tests
│   ├── encoder/                # Data encoding utilities
│   │   ├── encoder.go          # Data encoding utilities
│   │   └── encoder_test.go     # Encoder tests
│   ├── errors/                 # Error handling
│   │   └── types.go            # Error type definitions
│   ├── llm/                    # LLM integration
│   │   ├── llm.go              # AI service integration
│   │   └── llm_test.go         # LLM integration tests
│   ├── processor/              # Core processing logic
│   │   ├── processor.go        # Main processor coordination
│   │   ├── directory_processor.go    # Directory processing logic
│   │   ├── file_scanner.go           # File scanning functionality
│   │   ├── image_processor.go        # Image processing
│   │   ├── index_generator.go        # Index generation
│   │   ├── processor_test.go         # Processor tests
│   │   ├── directory_processor_test.go
│   │   ├── file_scanner_test.go
│   │   ├── image_processor_test.go
│   │   └── integration_test.go
│   ├── utils/                  # Utility functions
│   │   ├── fs.go               # File system utilities
│   │   └── fs_test.go          # File system tests
│   └── web/                    # Web interface
│       ├── api/                # API layer
│       │   ├── handlers.go     # API handlers
│       │   ├── middleware.go   # API middleware
│       │   └── sorting.go      # Sorting utilities
│       ├── queue/              # Background processing
│       │   ├── queue.go        # Task queue implementation
│       │   ├── queue_test.go   # Queue tests
│       │   └── queue_integration_test.go
│       ├── server.go           # Web server setup
│       ├── services/           # Service layer
│       │   ├── catalog_service.go      # Catalog service
│       │   └── template_renderer.go    # Template rendering
│       └── watch/              # File monitoring
│           ├── watch.go        # File system watcher
│           └── watch_test.go   # Watcher tests
├── web/                        # Web resources
│   ├── static/                 # Static assets
│   │   ├── htmx.min.js        # HTMX library
│   │   ├── styles.css          # CSS styles
│   │   ├── viewer.min.css      # Image viewer CSS
│   │   └── viewer.min.js       # Image viewer JS
│   ├── templates/              # HTML templates
│   │   ├── index.html                       # Main index page
│   │   ├── catalog-detail.html              # Catalog detail page
│   │   ├── catalog-images-fragment.html      # Images fragment
│   │   ├── catalog-images-template.html     # Images template
│   │   ├── catalog-list-fragment.html       # List fragment
│   │   ├── catalog-list-template.html       # List template
│   │   ├── catalog-navigation-template.html  # Navigation template
│   │   ├── search-result-fragment.html      # Search fragment
│   │   └── search-result.html              # Search results
│   ├── webresources.go         # Web resource embedding
│   └── webresources_test.go    # Web resource tests
├── docs/                       # Documentation
│   └── adr/                    # Architecture Decision Records
│       ├── ADR-0001-record-architecture-decisions.md
│       ├── ADR-0002-use-go-for-backend.md
│       ├── ADR-0003-htmx-for-web-interface.md
│       ├── ADR-0004-llm-integration.md
│       ├── ADR-0005-project-structure.md
│       ├── ADR-0006-error-handling.md
│       ├── ADR-0007-performance-optimization.md
│       ├── index.md
│       ├── README.md
│       └── template.md
├── config.yaml                 # Application configuration
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── LICENSE                     # Project license (AGPLv3)
├── .gitignore                  # Git ignore file
├── .log4brains.yml             # Log4brains configuration
├── Dockerfile                  # Docker image configuration
├── docker-compose.yml          # Docker Compose configuration
├── build-linux.sh              # Linux build script
├── build-win.sh                # Windows build script
└── README.md                   # Project documentation
```

### 📊 Project Statistics

- **Total Files**: 65+ files
- **Go Source Files**: 37 files
- **Test Files**: 15 files  
- **HTML Templates**: 9 files
- **Documentation**: 10+ files (including ADRs)
- **Build Scripts**: 2 files
- **Configuration Files**: 6 files

### 🏛️ Architecture Overview

The application follows a clean architecture pattern with clear separation of concerns:

- **CLI Layer** (`cmd/`): Command-line interface with graceful shutdown handling
- **Core Processing** (`internal/processor/`): Business logic for image analysis and catalog generation
- **AI Integration** (`internal/llm/`): Communication with LLM services
- **Web Interface** (`internal/web/`): HTTP server with HTMX-powered frontend
- **Configuration** (`internal/config/`): Centralized configuration management

### Technology Stack

#### 🔧 Core Technologies

- **Go 1.24+** - Primary programming language
- **Gin/Gorilla Mux** - HTTP routing (using standard http.Server)
- **HTMX** - Dynamic web interface without JavaScript frameworks
- **HTML/CSS** - Web interface with modern design

#### 🤖 AI/ML Integration

- **LLaVA v1.5-7b** - Image analysis model
- **Qwen-VL-8B-Instruct** - Alternative analysis model
- **OpenAI API Compatible** - Support for compatible APIs

#### 📦 Dependencies

```go
require (
    github.com/fsnotify/fsnotify v1.9.0     // File system monitoring
    golang.org/x/image v0.34.0             // Image processing
    gopkg.in/yaml.v2 v2.4.0               // YAML configuration parsing
)
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.24+** [Download Go](https://golang.org/dl/)
- **AI model** with compatible API (LLaVA, Qwen-VL, etc.)
- **Image collection** for cataloging
- **Docker and Docker Compose** (for containerized deployment)

### Using Docker Compose

To quickly deploy the application using Docker, use the provided docker-compose.yml file:

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

The application will be available at http://localhost:8080

### Installation

1. **Clone the repository:**

```bash
git clone <repository-url>
cd kbase-catalog
```

2. **Install dependencies:**

```bash
go mod tidy
```

3. **Configure the application:**
   Copy and configure `config.yaml`:

```yaml
api_url: "http://192.168.1.7:1234/v1/chat/completions"
model: "llava-v1.5-7b"
timeout: 60
system_prompt: |-
  You are a helpful assistant specialized in image analysis.
  You must respond in valid JSON format ONLY, without any extra text.
  The JSON must contain two keys:
  1. "short_name": a short, descriptive name for the image.
  2. "description": a detailed description of the image in English.

  Example output format:
  {"short_name": "Sunset on the beach", "description": "The image shows a sunset at sea..."}
supported_extensions:
  - ".png"
  - ".jpg"
  - ".jpeg"
  - ".webp"
  - ".gif"
  - ".bmp"
parallel_requests: 3
max_retries: 3
retry_delay: 5
```

### Usage

#### Command Line Interface

```bash
# Process entire catalog
go run cmd/kbase-catalog/main.go process /path/to/images

# Test single image
go run cmd/kbase-catalog/main.go test /path/to/image.jpg

# Start web interface
go run cmd/kbase-catalog/main.go web

# Start web interface with custom parameters
go run cmd/kbase-catalog/main.go -archive-dir /path/to/custom/archive -port 8080 web

# Show version
go run cmd/kbase-catalog/main.go version

# Show help
go run cmd/kbase-catalog/main.go
```

#### Build Binary

```bash
# Build for Linux
./build-linux.sh

# Build for Windows
./build-win.sh

# Manual build
go build -o kbase-catalog cmd/kbase-catalog/main.go
```

## 📊 Usage Examples

### Generated Metadata Example

```json
{
  "assassins-creed-origins-abilities-1920x1080.jpg": {
    "description": "A screenshot of a decision log entry from the 'tekBlueprint' architecture knowledge base, detailing the selection of XML as the blueprint format. The page outlines the context (need for a portable, widely-used format), considered options (JSON, YAML, XML, Custom DSL, TOML), and the rationale for choosing XML due to its commonality, broad editor support, and strong typing capabilities with XSD.",
    "original_name": "log4brains.png",
    "short_name": "Blueprint Format Decision",
    "update_date": "2026-01-08T13:55:56+04:00",
    "vl_model": "qwen3-vl-8b-instruct"
  }
}
```

### Web Interface

- **Home Page** - List of all catalogs with sorting capabilities
- **Catalog Page** - View all images in a catalog with search functionality
- **Search Results** - Global search across the entire collection
- **Auto-refresh** - Interface updates automatically when new files are added

## 🔧 Configuration

### Configuration Parameters

| Parameter              | Type     | Default                                | Description                     |
|------------------------|----------|----------------------------------------|---------------------------------|
| `api_url`              | string   | -                                      | AI API endpoint URL             |
| `model`                | string   | -                                      | Model name for analysis         |
| `timeout`              | int      | 60                                     | Request timeout in seconds      |
| `parallel_requests`    | int      | 3                                      | Number of parallel requests     |
| `max_retries`          | int      | 3                                      | Maximum retry attempts          |
| `retry_delay`          | int      | 5                                      | Delay between retries (seconds) |
| `supported_extensions` | []string | [.png, .jpg, .jpeg, .webp, .gif, .bmp] | Supported file formats          |

## 🧪 Testing and Development

### Test Structure

The project uses standard Go testing approaches:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/processor
```

### Code Quality

```bash
# Format code
go fmt ./...

# Linting
go vet ./...

# Error checking
golangci-lint run
```

## 🐳 Docker Deployment

A docker-compose.yml file has been provided to easily deploy both the KBase Image Catalog application and its required
AI model service. The setup includes:

1. **KBase Image Catalog Application**: Runs on port 8080
2. **AI Model Service**: LLaVA v1.5-7b running on port 1234

### Usage

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

> Note: Make sure to configure the `config.yaml` file with the appropriate AI model endpoint URL (should be
`http://ai-model:1234/v1/chat/completions` when using Docker Compose).

## 🚀 Deployment

### Docker (Recommended)

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kbase-catalog .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/kbase-catalog .
COPY --from=builder /app/config.yaml .
EXPOSE 8080
CMD ["./kbase-catalog", "web", "-port", "8080"]
```

### systemd Service

```ini
[Unit]
Description=KBase Image Catalog Service
After=network.target

[Service]
Type=simple
User=kbase
WorkingDirectory=/opt/kbase-catalog
ExecStart=/opt/kbase-catalog/kbase-catalog web -port 8080
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## 📈 Monitoring and Logging

### 📊 Log Structure

The application provides comprehensive logging with detailed information:

- **📁 Catalog Processing** - Progress tracking, statistics, and completion metrics
- **🤖 AI Requests** - Execution time, success/failure rates, and error details
- **🌐 Web Server** - HTTP request/response logging and error tracking
- **🔄 File Monitoring** - Real-time file system events and change notifications

### ⚡ Performance Metrics

Monitor system performance with these key indicators:

- **📈 Processing Speed** - Images processed per minute (IPM)
- **🎯 AI Success Rate** - Percentage of successful AI analyses
- **⏱️ Response Time** - Average time per image processing
- **💾 Memory Usage** - Current memory consumption and optimization

### 📋 Logging Examples

```bash
# View application logs
tail -f /var/log/kbase-catalog/app.log

# Monitor AI processing performance
grep "AI_REQUEST" /var/log/kbase-catalog/app.log | grep -E "success|failed"

# Track file system events
grep "FILE_EVENT" /var/log/kbase-catalog/app.log
```

## 🔒 Security

### 🛡️ Security Practices

The application implements comprehensive security measures to ensure safe operation:

- **🔍 Input Validation** - Rigorous file path and input validation to prevent directory traversal
- **⚙️ Resource Limitation** - Configurable parallel request limits to prevent resource exhaustion
- **🛡️ Error Handling** - Safe exception handling with detailed logging without information leakage
- **🔒 CORS Protection** - Web interface security headers and cross-origin protection
- **🧹 Input Sanitization** - Comprehensive user input cleaning and validation
- **🔐 File System Security** - Restricted file system access with proper permissions

### 🔍 Security Considerations

- **File Access**: Application only accesses explicitly configured directories
- **Network Security**: All AI API communications use secure endpoints (HTTPS recommended)
- **Memory Safety**: Go's memory safety features prevent buffer overflow vulnerabilities
- **Configuration Security**: Sensitive configuration stored in secure, restricted-access files

## 🤝 Contributing

We welcome and appreciate community contributions! Please follow our development guidelines to ensure high-quality
contributions.

### 📝 Development Guidelines

#### 🔄 Contribution Process

1. **Fork the repository** from the main project
2. **Create a feature branch** for your contribution: `git checkout -b feature/amazing-feature`
3. **Implement your changes** following our coding standards
4. **Add comprehensive tests** for new functionality
5. **Commit your changes**: `git commit -m 'feat: Add amazing feature'`
6. **Push to your branch**: `git push origin feature/amazing-feature`
7. **Create a Pull Request** with detailed description of your changes

### 📋 Code Requirements

#### 🎯 Quality Standards

- **📏 Go Conventions**: Follow [Go Programming Language conventions](https://golang.org/doc/effective_go.html) and our
  existing code patterns
- **🧪 Test Coverage**: Add comprehensive unit tests for new functionality (aim for >80% coverage)
- **📚 Documentation**: Update documentation when changing APIs or adding new features
- **🔍 Code Review**: Ensure code passes all quality checks before submission

#### 🔍 Quality Assurance

Before submitting your PR:

```bash
# Format code
go fmt ./...

# Run tests
go test -cover ./...

# Vet for issues
go vet ./...

# Lint code (if available)
golangci-lint run
```

### 📚 Documentation

- **API Changes**: Update documentation for any public API changes
- **README Updates**: Add or update README sections for new features
- **Code Comments**: Add inline comments for complex logic
- **Examples**: Provide usage examples for new functionality

### 🎯 Pull Request Guidelines

- **Title**: Use [conventional commits](https://www.conventionalcommits.org/) format
- **Description**: Provide detailed description of changes and their impact
- **Tests**: Include automated tests for new functionality
- **Documentation**: Update relevant documentation
- **Breaking Changes**: Clearly indicate any breaking changes

## 📄 License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPLv3)**.

### 📋 License Summary

- **Commercial Use**: ✅ Permitted
- **Modification**: ✅ Permitted
- **Distribution**: ✅ Permitted
- **Private Use**: ✅ Permitted
- **Sublicensing**: ❌ Not permitted
- **Patent Use**: ❌ Not permitted
- **Liability**: ❌ Limited
- **Warranty**: ❌ No warranty provided

### 📄 Full License

See the [LICENSE](LICENSE) file for the full license text.

### 🤔 Why AGPLv3?

We chose AGPLv3 to ensure that any modifications or enhancements made to the software remain free and open source, even
when used in networked services. This promotes:

- 🔄 **Open Source Ecosystem**: Ensures continuous improvement and sharing
- 🌐 **Network Freedom**: Protects users when software is used via network services
- 🤝 **Community Collaboration**: Encourages contribution and sharing of improvements

### 📝 Additional Terms

If you develop new programs that use this software, you should:

1. **Include License Notice**: Prominently display the AGPLv3 license
2. **Provide Source**: Make the corresponding source code available
3. **State Changes**: Clearly mark any modifications made to the original code
4. **Include Warranty Disclaimer**: Include the standard AGPLv3 warranty disclaimer

## 🆘 Support & Community

### 📖 Documentation

- **📚 Main Documentation**: This README file
- **🏗️ Architecture Records**: [docs/adr/](docs/adr/) - Architecture Decision Records
- **🔧 API Reference**: Inline code documentation and examples
- **🚀 Deployment Guides**: Docker, systemd, and cloud deployment instructions

### ❓ Frequently Asked Questions

**🤖 AI Model Selection**
> **Q: Which AI model should I use?**
>
> **A**: We recommend LLaVA v1.5-7b or Qwen-VL-8B-Instruct for better analysis quality. Consider your hardware
> capabilities and required processing speed.

**⚡ Performance Optimization**
> **Q: How to optimize performance?**
>
> **A**:
> - Increase `parallel_requests` in config.yaml (3-5 for most systems)
> - Use SSD storage for faster image access
> - Ensure stable network connection to AI service
> - Monitor system resources and adjust accordingly

**📁 Format Support**
> **Q: Can I add support for new image formats?**
>
> **A**: Yes, add the file extension to `supported_extensions` array in config.yaml. The system will automatically
> recognize and process files with those extensions.

**🐳 Docker Deployment**
> **Q: How to deploy with Docker Compose?**
>
> **A**: Use the provided docker-compose.yml file. Ensure AI model service is running and configured properly in
> config.yaml.

**🌐 Web Interface**
> **Q: Can I customize the web interface?**
>
> **A**: Yes, modify files in `web/templates/` and `web/static/` directories. The interface uses HTMX for dynamic
> updates.

### 💬 Community Support

#### 🐛 Bug Reports

- **GitHub Issues**: [Report bugs and request features](https://github.com/your-repo/kbase-catalog/issues)
- **Bug Template**: Use the provided issue template for structured bug reports

#### 💡 Feature Requests

- **GitHub Discussions**: [Share ideas and discuss features](https://github.com/your-repo/kbase-catalog/discussions)
- **Voting**: Vote on existing feature requests to prioritize development

#### 🤔 General Questions

- **GitHub Discussions**: [Ask questions and get help](https://github.com/devmix/kbase-image-catalog-ai/discussions)
- **Wiki**: Community knowledge base and tutorials

### 📧 Contact Information

- **🐛 Bug Reports**: [GitHub Issues](https://github.com/devmix/kbase-image-catalog-ai/issues)
- **💡 Feature Requests**: [GitHub Discussions](https://github.com/devmix/kbase-image-catalog-ai/discussions)

### 📊 Performance Monitoring

For production deployment, consider monitoring:

```bash
# Monitor processing performance
watch -n 5 "grep 'PROGRESS' /var/log/kbase-catalog/app.log | tail -10"

# Check AI service health
curl -f http://localhost:1234/health || echo "AI service unavailable"
```

## 🎯 Roadmap & Future Development

### 🚀 Planned Improvements (v0.2.0)

#### 🏗️ Core Enhancements

- [ ] **Multi-tasking Support** - Simultaneous processing of multiple catalogs
- [ ] **Cluster Processing** - Distributed processing for large collections
- [ ] **Integration API** - RESTful API for external applications
- [ ] **Advanced Search** - Search by tags, colors, objects, and AI-generated metadata

#### 🔧 Infrastructure

- [ ] **Cloud Storage Integration** - AWS S3, Google Cloud Storage, Azure Blob
- [ ] **Database Backend** - PostgreSQL/MySQL for metadata persistence
- [ ] **Caching Layer** - Redis for improved performance
- [ ] **Load Balancing** - Multiple instance support

#### 🌐 User Experience

- [ ] **Mobile Application** - React Native client for browsing catalogs
- [ ] **Desktop GUI** - Electron-based desktop application
- [ ] **Browser Extension** - Quick image analysis from web browsers
- [ ] **Multiple Export Formats** - PDF, CSV, XML, JSON catalog exports

#### 🤖 AI/ML Enhancements

- [ ] **Multiple Model Support** - Automatic model selection based on content type
- [ ] **Fine-tuned Models** - Custom models for specific domains (technical schemas, art, photos)
- [ ] **Batch Processing** - Optimized processing for large batches
- [ ] **Local Model Support** - Offline processing capabilities

### 📊 Version History

#### ✅ v0.1.0 (Current)

- **Core Features**: Image analysis, catalog generation, web interface
- **AI Integration**: LLaVA, Qwen-VL model support
- **CLI Interface**: Process, test, web commands
- **Docker Support**: Containerized deployment
- **Web Interface**: HTMX-powered, responsive design

#### 🚀 v0.2.0 (Planned)

- **API Layer**: RESTful API for external integrations
- **Performance**: Cluster processing and distributed computing
- **Storage**: Cloud storage integration
- **Mobile**: Cross-platform mobile applications

#### 🔮 v0.3.0 (Future Vision)

- **AI Assistant**: Interactive AI-powered catalog management
- **Machine Learning**: Automated tagging and categorization
- **Analytics**: Advanced usage statistics and insights
- **Enterprise**: Multi-user support and administration

### 🤝 Community Involvement

We welcome community input on our roadmap! Please:

1. **Vote on Issues**: Help prioritize features by voting on GitHub issues
2. **Submit Ideas**: Share your requirements in GitHub Discussions
3. **Contribute Code**: Implement features and submit pull requests
4. **Report Bugs**: Help us improve stability and performance

---

<div align="center">

**KBase Image Catalog** - Transform your image collections into intelligent catalogs with AI! 🚀✨

[🔝 Back to Top](#kbase-image-catalog---ai-powered-image-knowledge-management-system)

</div>