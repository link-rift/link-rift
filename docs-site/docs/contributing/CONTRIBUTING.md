# Contributing to Linkrift

> Last Updated: 2025-01-24

Thank you for your interest in contributing to Linkrift! This document provides guidelines and information for contributors.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Development Setup](#development-setup)
  - [Project Structure](#project-structure)
- [How to Contribute](#how-to-contribute)
  - [Reporting Issues](#reporting-issues)
  - [Suggesting Features](#suggesting-features)
  - [Contributing Code](#contributing-code)
- [Development Workflow](#development-workflow)
  - [Branching Strategy](#branching-strategy)
  - [Commit Guidelines](#commit-guidelines)
  - [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
  - [Go Style Guide](#go-style-guide)
  - [TypeScript/React Style Guide](#typescriptreact-style-guide)
  - [Documentation Standards](#documentation-standards)
- [Testing Requirements](#testing-requirements)
- [Review Process](#review-process)
- [Release Process](#release-process)
- [Community](#community)

---

## Code of Conduct

### Our Pledge

We pledge to make participation in our project a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.

### Our Standards

**Positive behaviors include:**
- Using welcoming and inclusive language
- Being respectful of differing viewpoints
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

**Unacceptable behaviors include:**
- Use of sexualized language or imagery
- Trolling, insulting/derogatory comments, and personal attacks
- Public or private harassment
- Publishing others' private information without permission
- Other conduct which could be considered inappropriate

### Enforcement

Instances of abusive, harassing, or otherwise unacceptable behavior may be reported by contacting the project team at conduct@linkrift.io. All complaints will be reviewed and investigated promptly.

---

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.22+ | Backend development |
| Node.js | 20+ | Frontend development |
| Docker | 24+ | Local development environment |
| Docker Compose | 2.20+ | Container orchestration |
| Git | 2.40+ | Version control |
| Make | 4.0+ | Build automation |

**Recommended tools:**
- [golangci-lint](https://golangci-lint.run/) - Go linting
- [air](https://github.com/cosmtrek/air) - Go hot reload
- [pnpm](https://pnpm.io/) - Fast package manager (optional)

### Development Setup

1. **Fork and Clone**

   ```bash
   # Fork the repository on GitHub, then clone your fork
   git clone https://github.com/YOUR_USERNAME/linkrift.git
   cd linkrift

   # Add upstream remote
   git remote add upstream https://github.com/linkrift/linkrift.git
   ```

2. **Install Dependencies**

   ```bash
   # Install Go dependencies
   go mod download

   # Install Node.js dependencies
   cd frontend && npm install && cd ..

   # Install development tools
   make install-tools
   ```

3. **Set Up Environment**

   ```bash
   # Copy environment template
   cp .env.example .env

   # Edit with your local settings
   vim .env
   ```

4. **Start Development Services**

   ```bash
   # Start PostgreSQL and Redis
   docker-compose up -d postgres redis

   # Run database migrations
   make migrate-up

   # Seed development data (optional)
   make seed
   ```

5. **Run the Application**

   ```bash
   # Terminal 1: Start backend with hot reload
   make dev-backend

   # Terminal 2: Start frontend with hot reload
   make dev-frontend

   # Or start everything at once
   make dev
   ```

6. **Verify Setup**

   ```bash
   # Run tests
   make test

   # Check linting
   make lint

   # Open the app
   open http://localhost:5173
   ```

### Project Structure

```
linkrift/
├── cmd/                    # Application entry points
│   ├── server/            # Main API server
│   └── worker/            # Background job worker
├── internal/              # Private application code
│   ├── api/              # HTTP handlers
│   ├── domain/           # Business logic
│   ├── repository/       # Data access
│   └── service/          # Application services
├── pkg/                   # Public packages
│   ├── shortcode/        # Short code generation
│   └── validator/        # Input validation
├── frontend/              # React frontend
│   ├── src/
│   │   ├── components/   # React components
│   │   ├── hooks/        # Custom hooks
│   │   ├── pages/        # Page components
│   │   └── utils/        # Utilities
│   └── tests/            # Frontend tests
├── migrations/            # Database migrations
├── scripts/               # Development scripts
├── docs/                  # Documentation
└── docker/               # Docker configurations
```

---

## How to Contribute

### Reporting Issues

Before creating an issue, please:

1. **Search existing issues** to avoid duplicates
2. **Check the FAQ** for common questions
3. **Use the latest version** to ensure the issue still exists

**When creating an issue:**

```markdown
## Description
Clear description of the issue

## Steps to Reproduce
1. Go to '...'
2. Click on '...'
3. See error

## Expected Behavior
What you expected to happen

## Actual Behavior
What actually happened

## Environment
- OS: [e.g., macOS 14.0]
- Browser: [e.g., Chrome 120]
- Linkrift Version: [e.g., 1.2.0]

## Additional Context
Screenshots, logs, or other relevant information
```

### Suggesting Features

Feature requests are welcome! Please:

1. **Search existing issues** for similar requests
2. **Be specific** about the problem you're solving
3. **Provide context** on why this feature would be useful

**Feature request template:**

```markdown
## Problem Statement
Describe the problem or need this feature addresses

## Proposed Solution
Your idea for solving this problem

## Alternatives Considered
Other solutions you've considered

## Additional Context
Mockups, examples, or references
```

### Contributing Code

We welcome code contributions! Here's the process:

1. **Find or create an issue** describing your contribution
2. **Comment on the issue** to express interest
3. **Wait for assignment** before starting work
4. **Create a branch** following our naming convention
5. **Write code** following our coding standards
6. **Add tests** for new functionality
7. **Submit a pull request** with a clear description

---

## Development Workflow

### Branching Strategy

We use a simplified Git Flow:

```
main ─────────────────────────────────────────> stable releases
  │
  └── feature/add-custom-domains ──────────────> feature branches
  └── fix/redirect-loop ───────────────────────> bug fixes
  └── docs/api-reference ──────────────────────> documentation
  └── refactor/optimize-queries ───────────────> refactoring
```

**Branch naming conventions:**

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

**Types:**

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Code style (formatting, semicolons) |
| `refactor` | Code change that neither fixes nor adds |
| `perf` | Performance improvement |
| `test` | Adding or fixing tests |
| `chore` | Maintenance tasks |

**Examples:**

```bash
# Feature
feat(api): add custom domain support

# Bug fix
fix(redirect): handle expired links correctly

# With scope and body
feat(analytics): implement real-time click tracking

- Add WebSocket connection for live updates
- Implement click aggregation service
- Add Redis pub/sub for event broadcasting

Closes #123

# Breaking change
feat(api)!: change authentication to OAuth2

BREAKING CHANGE: API now requires OAuth2 tokens instead of API keys.
See migration guide: docs/migration/v2.md
```

### Pull Request Process

1. **Before submitting:**
   - Ensure all tests pass (`make test`)
   - Run linters (`make lint`)
   - Update documentation if needed
   - Rebase on latest main branch

2. **PR title** should follow conventional commits:
   ```
   feat(analytics): add geographic breakdown chart
   ```

3. **PR description template:**

   ```markdown
   ## Summary
   Brief description of changes

   ## Changes
   - Change 1
   - Change 2

   ## Related Issues
   Closes #123

   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Integration tests added/updated
   - [ ] Manual testing completed

   ## Screenshots (if applicable)
   Before/after screenshots for UI changes

   ## Checklist
   - [ ] Code follows project style guidelines
   - [ ] Self-review completed
   - [ ] Documentation updated
   - [ ] No breaking changes (or documented)
   ```

4. **After submission:**
   - Address review comments promptly
   - Keep the PR focused and small
   - Don't force-push after review starts

---

## Coding Standards

### Go Style Guide

We follow the [Effective Go](https://golang.org/doc/effective_go) guidelines with additional conventions:

**Package organization:**

```go
// Good: Clear package purpose
package shortcode

// Good: Package comment
// Package shortcode provides URL shortening functionality.
package shortcode
```

**Error handling:**

```go
// Good: Wrap errors with context
if err := db.Save(link); err != nil {
    return fmt.Errorf("saving link %s: %w", link.ID, err)
}

// Good: Custom errors
var ErrLinkNotFound = errors.New("link not found")

// Good: Error checking
if errors.Is(err, ErrLinkNotFound) {
    return http.StatusNotFound
}
```

**Naming conventions:**

```go
// Good: Descriptive names
func GenerateShortCode(length int) (string, error)

// Good: Interface naming
type LinkRepository interface {
    Save(link *Link) error
    FindByCode(code string) (*Link, error)
}

// Good: Constants
const (
    DefaultCodeLength = 6
    MaxCodeLength     = 12
)
```

**Testing:**

```go
// Good: Table-driven tests
func TestValidateURL(t *testing.T) {
    tests := []struct {
        name  string
        url   string
        valid bool
    }{
        {"valid http", "http://example.com", true},
        {"empty", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### TypeScript/React Style Guide

We follow the [React TypeScript Cheatsheet](https://react-typescript-cheatsheet.netlify.app/) conventions:

**Component structure:**

```tsx
// Good: Functional component with TypeScript
interface LinkFormProps {
  onSubmit: (url: string) => void;
  isLoading?: boolean;
}

export function LinkForm({ onSubmit, isLoading = false }: LinkFormProps) {
  const [url, setUrl] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(url);
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* ... */}
    </form>
  );
}
```

**Hooks:**

```tsx
// Good: Custom hook with proper typing
function useLinks() {
  return useQuery<Link[], Error>({
    queryKey: ['links'],
    queryFn: fetchLinks,
  });
}
```

**File organization:**

```
components/
├── LinkForm/
│   ├── LinkForm.tsx        # Component
│   ├── LinkForm.test.tsx   # Tests
│   ├── LinkForm.styles.ts  # Styles (if applicable)
│   └── index.ts            # Re-export
```

### Documentation Standards

**Code comments:**

```go
// GenerateShortCode creates a cryptographically random short code
// of the specified length using base62 encoding.
//
// The length must be between 4 and 12 characters.
// Returns an error if the length is invalid or random generation fails.
func GenerateShortCode(length int) (string, error) {
```

**API documentation:**

```go
// CreateLink creates a new shortened URL.
//
//	@Summary      Create a short link
//	@Description  Creates a new shortened URL with optional custom code
//	@Tags         links
//	@Accept       json
//	@Produce      json
//	@Param        request body CreateLinkRequest true "Link creation request"
//	@Success      201 {object} Link
//	@Failure      400 {object} ErrorResponse
//	@Router       /api/v1/links [post]
func (h *LinkHandler) Create(w http.ResponseWriter, r *http.Request) {
```

---

## Testing Requirements

All contributions must include appropriate tests:

| Change Type | Required Tests |
|-------------|----------------|
| New feature | Unit + integration tests |
| Bug fix | Regression test |
| API change | API tests + documentation |
| UI change | Component tests |
| Performance | Benchmark tests |

**Minimum coverage requirements:**
- New code: 80% line coverage
- Overall project: 75% line coverage

**Running tests:**

```bash
# All tests
make test

# Backend only
make test-backend

# Frontend only
make test-frontend

# With coverage
make coverage

# Specific package
go test -v ./internal/service/...

# Watch mode (frontend)
npm run test -- --watch
```

---

## Review Process

All pull requests require:

1. **Automated checks passing:**
   - CI build succeeds
   - All tests pass
   - Linting passes
   - Coverage requirements met

2. **Code review approval:**
   - At least one maintainer approval
   - All review comments addressed
   - No unresolved discussions

3. **Review criteria:**
   - Code quality and readability
   - Test coverage and quality
   - Documentation accuracy
   - Security considerations
   - Performance impact

**Review timeline:**
- Initial review: within 3 business days
- Follow-up reviews: within 1 business day

---

## Release Process

Releases follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking API changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

**Release checklist:**

1. Update `CHANGELOG.md`
2. Update version numbers
3. Create release branch
4. Run full test suite
5. Create GitHub release
6. Deploy to staging
7. Verify staging
8. Deploy to production

---

## Community

**Get help:**
- [GitHub Discussions](https://github.com/linkrift/linkrift/discussions) - Questions and ideas
- [Discord](https://discord.gg/linkrift) - Real-time chat
- [Stack Overflow](https://stackoverflow.com/questions/tagged/linkrift) - Technical questions

**Stay updated:**
- [Blog](https://linkrift.io/blog) - Announcements and tutorials
- [Twitter](https://twitter.com/linkrift) - Updates and news
- [Newsletter](https://linkrift.io/newsletter) - Monthly digest

**Recognition:**
- Contributors are listed in [CONTRIBUTORS.md](../CONTRIBUTORS.md)
- Significant contributions are highlighted in release notes
- Top contributors may be invited to join the maintainers team

---

Thank you for contributing to Linkrift! Your efforts help make this project better for everyone.
