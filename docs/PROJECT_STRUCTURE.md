# Project Structure

```
mekhanikube/
├── .devcontainer/          # VS Code Dev Container configuration
│   └── devcontainer.json   # Container development environment
├── .github/                # GitHub configuration
│   └── workflows/          # CI/CD workflows
│       └── docker-build.yml # Build, test, and release pipeline
├── configs/                # Configuration files
│   ├── Dockerfile          # K8sGPT container image
│   └── entrypoint.sh       # Container startup script
├── docs/                   # Documentation
│   ├── ARCHITECTURE.md     # System architecture details
│   ├── FAQ.md              # Frequently asked questions
│   └── TROUBLESHOOTING.md  # Common issues and solutions
├── scripts/                # Utility scripts
│   ├── analyze.sh          # Quick analysis script
│   ├── change-model.sh     # Model switching script
│   ├── healthcheck.sh      # Health check script
│   ├── release.sh          # Release automation
│   └── test.sh             # Integration tests
├── tests/                  # Test files
│   └── (future test files)
├── .env.example            # Environment variables template
├── .gitignore              # Git ignore rules
├── CHANGELOG.md            # Version history
├── CODE_OF_CONDUCT.md      # Community guidelines
├── CONTRIBUTING.md         # Contribution guidelines
├── docker-compose.yml      # Service orchestration
├── LICENSE                 # MIT License
├── Makefile                # Build automation
├── README.md               # Main documentation
├── SECURITY.md             # Security policy
└── VERSION                 # Current version number
```

## Directory Purposes

### Root Level Files

- **README.md**: Main entry point with quick start and overview
- **Makefile**: Convenient commands for all operations
- **docker-compose.yml**: Service definitions and orchestration
- **VERSION**: Semantic version number
- **.env.example**: Template for environment configuration

### `.devcontainer/`

Development container configuration for VS Code:
- Pre-configured environment
- Includes Docker, kubectl, and necessary tools
- Consistent development setup across team

### `.github/`

GitHub-specific configuration:
- **workflows/**: CI/CD automation
  - Build and test on push
  - Security scanning
  - Release automation

### `configs/`

Core configuration files for containers:
- **Dockerfile**: Multi-stage build for K8sGPT
- **entrypoint.sh**: Container initialization logic

### `docs/`

Comprehensive documentation:
- **ARCHITECTURE.md**: System design and components
- **FAQ.md**: Common questions and answers
- **TROUBLESHOOTING.md**: Problem-solving guide

### `scripts/`

Utility scripts for common operations:
- **analyze.sh**: Simplified analysis
- **change-model.sh**: Model management
- **healthcheck.sh**: System diagnostics
- **release.sh**: Version release automation
- **test.sh**: Integration testing

### `tests/`

Test files and fixtures (for future expansion):
- Unit tests
- Integration tests
- End-to-end tests

## File Descriptions

### Configuration Files

| File | Purpose |
|------|---------|
| `.env.example` | Environment variable template |
| `.gitignore` | Files to exclude from git |
| `docker-compose.yml` | Container orchestration |
| `Makefile` | Task automation |

### Documentation Files

| File | Purpose |
|------|---------|
| `README.md` | Main documentation |
| `CHANGELOG.md` | Version history |
| `CONTRIBUTING.md` | How to contribute |
| `CODE_OF_CONDUCT.md` | Community standards |
| `SECURITY.md` | Security policy |
| `LICENSE` | MIT license terms |

### Container Files

| File | Purpose |
|------|---------|
| `configs/Dockerfile` | K8sGPT image build |
| `configs/entrypoint.sh` | Container startup |

### Automation Files

| File | Purpose |
|------|---------|
| `.github/workflows/docker-build.yml` | CI/CD pipeline |
| `Makefile` | Build automation |
| `scripts/*.sh` | Utility scripts |

## Key Design Decisions

### 1. Separation of Concerns

- **configs/**: Container-specific files
- **scripts/**: Reusable automation
- **docs/**: User-facing documentation

### 2. Makefile as Primary Interface

Users interact via `make` commands rather than remembering docker commands:
```bash
make setup    # Instead of: docker-compose build && docker-compose up && ...
make analyze  # Instead of: docker exec mekhanikube-k8sgpt k8sgpt analyze --explain
```

### 3. Environment Flexibility

`.env.example` provides template while `.env` (git-ignored) allows customization.

### 4. Comprehensive Documentation

Multiple docs for different audiences:
- Beginners → README.md
- Troubleshooters → TROUBLESHOOTING.md
- Curious users → FAQ.md
- Architects → ARCHITECTURE.md

### 5. Developer Experience

- `.devcontainer/` for consistent dev environment
- `scripts/` for common operations
- `Makefile` for task automation
- CI/CD for quality assurance

## Adding New Components

### New Script

1. Create in `scripts/` directory
2. Make executable: `chmod +x scripts/new-script.sh`
3. Add to Makefile if needed
4. Document in README.md

### New Documentation

1. Create in `docs/` directory
2. Link from README.md
3. Update table of contents

### New Test

1. Create in `tests/` directory
2. Add to `scripts/test.sh`
3. Include in CI/CD pipeline

### New Configuration

1. Add to `configs/` directory
2. Reference in `docker-compose.yml`
3. Document in relevant docs

## File Naming Conventions

- **Markdown files**: UPPERCASE.md (e.g., README.md)
- **Config files**: lowercase or .extension (e.g., Dockerfile, .env)
- **Scripts**: lowercase-with-dashes.sh (e.g., health-check.sh)
- **Docs**: UPPERCASE.md in root, lowercase.md in docs/

## Git Workflow

```
main (protected)
  ↑
  └─ Pull Requests (require review)
       ↑
       └─ Feature branches
```

## Version Control

- **VERSION file**: Single source of truth
- **CHANGELOG.md**: Human-readable history
- **Git tags**: `v1.0.0` format
- **GitHub Releases**: Automated from tags
