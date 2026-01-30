# Contributing to Linkrift

Thank you for your interest in contributing to Linkrift! This document provides guidelines and information for contributors.

## Code of Conduct

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.

## Getting Started

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.22+ | Backend development |
| Node.js | 20+ | Frontend development |
| Docker | 24+ | Local development environment |
| Docker Compose | 2.20+ | Container orchestration |
| Git | 2.40+ | Version control |
| Make | 4.0+ | Build automation |

### Development Setup

1. **Fork and Clone**

   ```bash
   git clone https://github.com/YOUR_USERNAME/link-rift.git
   cd link-rift
   git remote add upstream https://github.com/link-rift/link-rift.git
   ```

2. **Install Dependencies**

   ```bash
   go mod download
   cd web && pnpm install && cd ..
   ```

3. **Set Up Environment**

   ```bash
   cp .env.example .env
   ```

4. **Start Development Services**

   ```bash
   docker compose up -d postgres redis
   make migrate-up
   ```

5. **Run the Application**

   ```bash
   make dev
   ```

6. **Verify Setup**

   ```bash
   make test
   make lint
   ```

## How to Contribute

### Reporting Issues

Before creating an issue:

1. Search existing issues to avoid duplicates
2. Check the FAQ for common questions
3. Use the latest version to ensure the issue still exists

### Contributing Code

1. Find or create an issue describing your contribution
2. Comment on the issue to express interest
3. Wait for assignment before starting work
4. Create a branch following our naming convention
5. Write code following our coding standards
6. Add tests for new functionality
7. Submit a pull request with a clear description

## Development Workflow

### Branching Strategy

| Prefix | Purpose | Example |
|--------|---------|---------|
| `feature/` | New features | `feature/qr-codes` |
| `fix/` | Bug fixes | `fix/analytics-timezone` |
| `docs/` | Documentation | `docs/api-v2` |
| `refactor/` | Code improvements | `refactor/service-layer` |
| `test/` | Test improvements | `test/e2e-coverage` |
| `chore/` | Maintenance | `chore/update-deps` |

### Commit Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`

**Examples:**

```bash
feat(api): add custom domain support
fix(redirect): handle expired links correctly
feat(analytics): implement real-time click tracking
```

### Pull Request Process

1. Ensure all tests pass (`make test`)
2. Run linters (`make lint`)
3. Update documentation if needed
4. Rebase on latest main branch

## Coding Standards

### Go Style Guide

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Wrap errors with context: `fmt.Errorf("saving link %s: %w", link.ID, err)`
- Use table-driven tests
- Run `golangci-lint` before submitting

### TypeScript/React Style Guide

- Follow [React TypeScript Cheatsheet](https://react-typescript-cheatsheet.netlify.app/) conventions
- Use functional components with TypeScript interfaces
- Use custom hooks for shared logic

## Testing Requirements

| Change Type | Required Tests |
|-------------|----------------|
| New feature | Unit + integration tests |
| Bug fix | Regression test |
| API change | API tests + documentation |
| UI change | Component tests |
| Performance | Benchmark tests |

**Minimum coverage:** 80% line coverage for new code.

```bash
make test          # All tests
make test-cover    # With coverage
```

## Review Process

All pull requests require:

1. Automated CI checks passing
2. At least one maintainer approval
3. All review comments addressed

## Community

- [GitHub Discussions](https://github.com/link-rift/link-rift/discussions) - Questions and ideas
- [GitHub Issues](https://github.com/link-rift/link-rift/issues) - Bug reports and feature requests

---

Thank you for contributing to Linkrift!
