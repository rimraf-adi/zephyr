# Zephyr - A Modern Python Package Manager

Zephyr is a fast, reliable Python package manager that uses the Pubgrub dependency resolution algorithm. It provides a modern alternative to pip with better dependency resolution, lockfile support, and a streamlined workflow.

## Features

- **Fast Dependency Resolution**: Uses the Pubgrub algorithm for efficient dependency solving
- **PyPI Integration**: Full compatibility with the Python Package Index
- **Virtual Environment Management**: Built-in virtual environment creation and management
- **Lockfile Support**: Deterministic builds with `zephyr.lock`
- **buildmeta.yaml**: Modern project configuration format
- **PEP Compliance**: Supports PEP 517, 518, and 621 standards
- **Wheel Installation**: Native wheel file handling and installation
- **Custom/Private Index Support**: Configure PyPI or any custom index via config file or environment variable.
- **Config Files**: Supports global (~/.zephyr/config.yaml) and project-level (.zephyrrc) configuration.

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/zephyr.git
cd zephyr

# Build the binary
go build -o zephyr cmd/zephyr/main.go

# Install globally (optional)
go install ./cmd/zephyr
```

## Quick Start

### 1. Initialize a new project

```bash
zephyr init my-python-project
```

This creates a `buildmeta.yaml` file with basic project configuration.

### 2. Add dependencies

```bash
zephyr add requests ">=2.25.0"
zephyr add flask ">=2.0.0"
```

### 3. Install dependencies

```bash
zephyr install
```

This resolves dependencies and creates a `zephyr.lock` file.

### 4. Create and use virtual environment

```bash
zephyr venv create
zephyr venv install
```

## Project Structure

```
zephyr/
├── cmd/zephyr/          # CLI application
├── pkg/
│   ├── solver/          # Pubgrub dependency solver
│   ├── pypi/            # PyPI API integration
│   ├── installer/       # Package installation logic
│   ├── buildmeta/       # buildmeta.yaml handling
│   └── netutil/         # HTTP and parsing utilities
├── buildmeta.yaml       # Project configuration
├── zephyr.lock          # Dependency lockfile
└── README.md
```

## Configuration

### buildmeta.yaml

The `buildmeta.yaml` file is the main configuration file for your project:

```yaml
version: "0.1.0"
name: "my-python-project"
description: "A Python project created with Zephyr"
author: "Your Name"
email: "your.email@example.com"
license: "MIT"

python:
  requires: ">=3.8"
  packages: ["my_package"]

build:
  backend: "setuptools.build_meta"

dependencies:
  direct:
    requests: ">=2.25.0"
    flask: ">=2.0.0"

dev-dependencies:
  direct:
    pytest: ">=6.0.0"
    black: ">=21.0.0"

optional-dependencies:
  test:
    pytest-cov: ">=2.0.0"
  dev:
    mypy: ">=0.900"

scripts:
  start: "python -m my_package"
  test: "pytest"

entry-points:
  console_scripts:
    my-app: "my_package.cli:main"
```

### Global and Project Config

- **Global config**: `~/.zephyr/config.yaml`
- **Project config**: `./.zephyrrc`
- **Environment variable**: `ZEPHYR_INDEX_URL`

Example `config.yaml` or `.zephyrrc`:

```yaml
index_url: "https://pypi.org"
```

To use a private index:

```yaml
index_url: "https://mycompany.com/pypi"
```

Or set the environment variable:

```bash
export ZEPHYR_INDEX_URL="https://mycompany.com/pypi"
```

## CLI Commands

### Project Management

- `zephyr init [project-name]` - Initialize a new Python project
- `zephyr add <package> [constraint]` - Add a dependency
- `zephyr install` - Install project dependencies
- `zephyr search <query>` - Search for packages on PyPI

### Virtual Environment

- `zephyr venv create [path]` - Create a new virtual environment
- `zephyr venv install [venv-path]` - Install dependencies into virtual environment

### Development

- `zephyr solve` - Solve dependencies using Pubgrub algorithm
- `zephyr demo` - Run Pubgrub algorithm demonstration
- `zephyr examples` - Show Pubgrub algorithm examples

## Dependency Resolution

Zephyr uses the Pubgrub algorithm for dependency resolution, which provides:

- **Completeness**: Always finds a solution if one exists
- **Efficiency**: Fast resolution even for large dependency graphs
- **Conflict Detection**: Clear error messages when conflicts occur
- **Deterministic Results**: Same input always produces the same output

### Example Resolution

```bash
$ zephyr solve
✅ Dependencies solved successfully!

Solution:
  requests == 2.31.0
  urllib3 == 2.0.7
  certifi == 2023.11.17
  charset-normalizer == 3.3.2
  idna == 3.6
```

## Lockfile

The `zephyr.lock` file ensures reproducible builds by locking exact versions:

```json
{
  "version": "1.0",
  "generated_at": "2024-01-15T10:30:00Z",
  "python": "3.11",
  "packages": {
    "requests": {
      "version": "2.31.0",
      "source": "pypi",
      "url": "https://pypi.org/pypi/requests/2.31.0/json",
      "hash": "sha256:..."
    }
  },
  "metadata": {
    "hash": "1234567890",
    "resolved_by": "zephyr",
    "resolved_at": "2024-01-15T10:30:00Z"
  }
}
```

## PyPI Integration

Zephyr provides full PyPI integration:

- **Package Search**: Search for packages and view metadata
- **Version Discovery**: Find available versions for packages
- **Wheel Download**: Download and install wheel files
- **Metadata Parsing**: Parse package metadata and dependencies

### Example Search

```bash
$ zephyr search requests
📦 requests 2.31.0
📝 Python HTTP for Humans.
👤 Author: Kenneth Reitz
🌐 Homepage: https://requests.readthedocs.io

Available versions:
  2.31.0
  2.30.0
  2.29.0
  ...
```

## Virtual Environment Management

Zephyr includes built-in virtual environment management:

```bash
# Create virtual environment
zephyr venv create .venv

# Install dependencies into virtual environment
zephyr venv install .venv

# Activate virtual environment
source .venv/bin/activate  # Linux/macOS
.venv\Scripts\activate     # Windows
```

## PEP Compliance

Zephyr supports modern Python packaging standards:

- **PEP 517**: Build backend interface
- **PEP 518**: Build system requirements
- **PEP 621**: Project metadata

This ensures compatibility with existing Python tooling and workflows.

## Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/rimraf-adi/zephyr.git
cd zephyr

# Install dependencies
go mod download

# Build
go build -o zephyr cmd/zephyr/main.go

# Run tests
go test ./...
```

### Project Structure

- `pkg/solver/`: Core Pubgrub dependency resolution algorithm
- `pkg/pypi/`: PyPI API client and metadata handling
- `pkg/installer/`: Package installation and virtual environment management
- `pkg/buildmeta/`: buildmeta.yaml configuration handling
- `pkg/netutil/`: HTTP client and parsing utilities
- `cmd/zephyr/`: CLI application using Cobra

### Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./pkg/solver

# Run with coverage
go test -cover ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- **Pubgrub Algorithm**: Based on the paper "Dependency Resolution with Pubgrub" by Natalie Weizenbaum
- **PyPI**: Python Package Index for package metadata and distribution
- **Cobra**: CLI framework for Go
- **Go Modules**: Dependency management for Go

## Roadmap

- [ ] Support for private package repositories
- [ ] Plugin system for custom build backends
- [ ] Integration with CI/CD systems
- [ ] Performance optimizations
- [ ] Additional package formats (conda, etc.)
- [ ] GUI interface
- [ ] Package publishing workflow 