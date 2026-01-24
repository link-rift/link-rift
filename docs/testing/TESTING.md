# Testing Strategy

> Last Updated: 2025-01-24

Comprehensive testing documentation for Linkrift URL shortener, covering backend Go testing, frontend React testing, and performance testing strategies.

---

## Table of Contents

- [Testing Philosophy](#testing-philosophy)
- [Test Pyramid](#test-pyramid)
- [Backend Testing (Go)](#backend-testing-go)
  - [Table-Driven Tests](#table-driven-tests)
  - [Using Testify](#using-testify)
  - [HTTP Testing with httptest](#http-testing-with-httptest)
  - [Integration Testing with testcontainers-go](#integration-testing-with-testcontainers-go)
  - [Benchmark Tests](#benchmark-tests)
- [Frontend Testing](#frontend-testing)
  - [Unit Testing with Vitest](#unit-testing-with-vitest)
  - [Component Testing with React Testing Library](#component-testing-with-react-testing-library)
  - [End-to-End Testing with Playwright](#end-to-end-testing-with-playwright)
- [Load Testing](#load-testing)
  - [k6 Load Testing](#k6-load-testing)
  - [Vegeta Load Testing](#vegeta-load-testing)
- [Coverage Reporting](#coverage-reporting)
- [Continuous Integration](#continuous-integration)

---

## Testing Philosophy

Linkrift follows a comprehensive testing strategy that emphasizes:

1. **Fast Feedback**: Unit tests run in milliseconds, providing immediate feedback
2. **Confidence**: Integration tests verify system behavior with real dependencies
3. **Performance**: Regular load testing ensures the system handles expected traffic
4. **Maintainability**: Clear test patterns make tests easy to understand and modify

## Test Pyramid

```
                    ┌───────────────┐
                    │   E2E Tests   │  ← Few, slow, high confidence
                    │  (Playwright) │
                    ├───────────────┤
                    │  Integration  │  ← Some, medium speed
                    │    Tests      │
                    │(testcontainers)│
            ┌───────┴───────────────┴───────┐
            │        Unit Tests             │  ← Many, fast, isolated
            │   (Go tests, Vitest, RTL)     │
            └───────────────────────────────┘
```

**Recommended Distribution:**
- Unit Tests: 70%
- Integration Tests: 20%
- E2E Tests: 10%

---

## Backend Testing (Go)

### Table-Driven Tests

Table-driven tests are the idiomatic Go testing pattern. They provide excellent test coverage with minimal code duplication.

```go
package shortener

import (
    "testing"
)

func TestGenerateShortCode(t *testing.T) {
    tests := []struct {
        name     string
        length   int
        wantLen  int
        wantErr  bool
    }{
        {
            name:    "default length",
            length:  6,
            wantLen: 6,
            wantErr: false,
        },
        {
            name:    "custom length",
            length:  8,
            wantLen: 8,
            wantErr: false,
        },
        {
            name:    "zero length returns error",
            length:  0,
            wantLen: 0,
            wantErr: true,
        },
        {
            name:    "negative length returns error",
            length:  -1,
            wantLen: 0,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := GenerateShortCode(tt.length)

            if (err != nil) != tt.wantErr {
                t.Errorf("GenerateShortCode() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if len(got) != tt.wantLen {
                t.Errorf("GenerateShortCode() length = %v, want %v", len(got), tt.wantLen)
            }
        })
    }
}

func TestValidateURL(t *testing.T) {
    tests := []struct {
        name    string
        url     string
        isValid bool
    }{
        {"valid http URL", "http://example.com", true},
        {"valid https URL", "https://example.com/path", true},
        {"valid URL with query", "https://example.com?foo=bar", true},
        {"empty URL", "", false},
        {"no protocol", "example.com", false},
        {"invalid protocol", "ftp://example.com", false},
        {"javascript URL", "javascript:alert(1)", false},
        {"data URL", "data:text/html,<h1>Hello</h1>", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := ValidateURL(tt.url); got != tt.isValid {
                t.Errorf("ValidateURL(%q) = %v, want %v", tt.url, got, tt.isValid)
            }
        })
    }
}
```

### Using Testify

[Testify](https://github.com/stretchr/testify) provides assertion functions and mocking capabilities.

```go
package shortener

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)

// Assertions with testify
func TestLinkService_Create(t *testing.T) {
    service := NewLinkService(mockRepo)

    link, err := service.Create("https://example.com")

    require.NoError(t, err, "Create should not return an error")
    assert.NotEmpty(t, link.ShortCode, "ShortCode should be generated")
    assert.Equal(t, "https://example.com", link.OriginalURL)
    assert.False(t, link.CreatedAt.IsZero(), "CreatedAt should be set")
}

// Mock repository
type MockLinkRepository struct {
    mock.Mock
}

func (m *MockLinkRepository) Save(link *Link) error {
    args := m.Called(link)
    return args.Error(0)
}

func (m *MockLinkRepository) FindByShortCode(code string) (*Link, error) {
    args := m.Called(code)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Link), args.Error(1)
}

func TestLinkService_Resolve(t *testing.T) {
    mockRepo := new(MockLinkRepository)
    service := NewLinkService(mockRepo)

    expectedLink := &Link{
        ShortCode:   "abc123",
        OriginalURL: "https://example.com",
    }

    mockRepo.On("FindByShortCode", "abc123").Return(expectedLink, nil)

    link, err := service.Resolve("abc123")

    require.NoError(t, err)
    assert.Equal(t, expectedLink.OriginalURL, link.OriginalURL)
    mockRepo.AssertExpectations(t)
}

// Test Suite
type LinkServiceTestSuite struct {
    suite.Suite
    service *LinkService
    repo    *MockLinkRepository
}

func (s *LinkServiceTestSuite) SetupTest() {
    s.repo = new(MockLinkRepository)
    s.service = NewLinkService(s.repo)
}

func (s *LinkServiceTestSuite) TestCreate() {
    s.repo.On("Save", mock.AnythingOfType("*Link")).Return(nil)

    link, err := s.service.Create("https://example.com")

    s.Require().NoError(err)
    s.NotEmpty(link.ShortCode)
}

func (s *LinkServiceTestSuite) TestResolveNotFound() {
    s.repo.On("FindByShortCode", "notfound").Return(nil, ErrLinkNotFound)

    _, err := s.service.Resolve("notfound")

    s.ErrorIs(err, ErrLinkNotFound)
}

func TestLinkServiceTestSuite(t *testing.T) {
    suite.Run(t, new(LinkServiceTestSuite))
}
```

### HTTP Testing with httptest

Go's `net/http/httptest` package provides utilities for testing HTTP handlers.

```go
package api

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCreateLinkHandler(t *testing.T) {
    handler := NewLinkHandler(mockService)

    tests := []struct {
        name           string
        requestBody    map[string]interface{}
        expectedStatus int
        expectedBody   map[string]interface{}
    }{
        {
            name: "valid request",
            requestBody: map[string]interface{}{
                "url": "https://example.com",
            },
            expectedStatus: http.StatusCreated,
            expectedBody: map[string]interface{}{
                "short_code": "abc123",
                "short_url":  "https://lnkr.ft/abc123",
            },
        },
        {
            name: "missing URL",
            requestBody: map[string]interface{}{},
            expectedStatus: http.StatusBadRequest,
            expectedBody: map[string]interface{}{
                "error": "url is required",
            },
        },
        {
            name: "invalid URL",
            requestBody: map[string]interface{}{
                "url": "not-a-url",
            },
            expectedStatus: http.StatusBadRequest,
            expectedBody: map[string]interface{}{
                "error": "invalid URL format",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            body, _ := json.Marshal(tt.requestBody)
            req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewReader(body))
            req.Header.Set("Content-Type", "application/json")

            rec := httptest.NewRecorder()
            handler.ServeHTTP(rec, req)

            assert.Equal(t, tt.expectedStatus, rec.Code)

            var response map[string]interface{}
            err := json.Unmarshal(rec.Body.Bytes(), &response)
            require.NoError(t, err)

            for key, expectedValue := range tt.expectedBody {
                assert.Equal(t, expectedValue, response[key])
            }
        })
    }
}

func TestRedirectHandler(t *testing.T) {
    handler := NewRedirectHandler(mockService)

    // Setup mock
    mockService.On("Resolve", "abc123").Return(&Link{
        OriginalURL: "https://example.com",
    }, nil)

    req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusMovedPermanently, rec.Code)
    assert.Equal(t, "https://example.com", rec.Header().Get("Location"))
}

// Testing with a full test server
func TestAPIIntegration(t *testing.T) {
    server := httptest.NewServer(NewRouter())
    defer server.Close()

    // Create a short link
    createBody := bytes.NewBufferString(`{"url": "https://example.com"}`)
    resp, err := http.Post(server.URL+"/api/v1/links", "application/json", createBody)
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    var result struct {
        ShortCode string `json:"short_code"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    // Verify redirect works
    client := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }

    redirectResp, err := client.Get(server.URL + "/" + result.ShortCode)
    require.NoError(t, err)
    defer redirectResp.Body.Close()

    assert.Equal(t, http.StatusMovedPermanently, redirectResp.StatusCode)
    assert.Equal(t, "https://example.com", redirectResp.Header.Get("Location"))
}
```

### Integration Testing with testcontainers-go

[Testcontainers-go](https://golang.testcontainers.org/) enables running real dependencies in Docker containers during tests.

```go
package integration

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/modules/redis"
)

// PostgreSQL Integration Test
func TestLinkRepository_PostgreSQL(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    ctx := context.Background()

    // Start PostgreSQL container
    pgContainer, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:16-alpine"),
        postgres.WithDatabase("linkrift_test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(5*time.Second),
        ),
    )
    require.NoError(t, err)
    defer pgContainer.Terminate(ctx)

    // Get connection string
    connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
    require.NoError(t, err)

    // Run migrations
    db, err := sql.Open("postgres", connStr)
    require.NoError(t, err)
    defer db.Close()

    err = RunMigrations(db)
    require.NoError(t, err)

    // Create repository and run tests
    repo := NewPostgresLinkRepository(db)

    t.Run("Save and FindByShortCode", func(t *testing.T) {
        link := &Link{
            ShortCode:   "test123",
            OriginalURL: "https://example.com",
            CreatedAt:   time.Now(),
        }

        err := repo.Save(link)
        require.NoError(t, err)

        found, err := repo.FindByShortCode("test123")
        require.NoError(t, err)
        require.Equal(t, link.OriginalURL, found.OriginalURL)
    })

    t.Run("FindByShortCode not found", func(t *testing.T) {
        _, err := repo.FindByShortCode("notfound")
        require.ErrorIs(t, err, ErrLinkNotFound)
    })
}

// Redis Integration Test
func TestCacheRepository_Redis(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    ctx := context.Background()

    // Start Redis container
    redisContainer, err := redis.RunContainer(ctx,
        testcontainers.WithImage("redis:7-alpine"),
    )
    require.NoError(t, err)
    defer redisContainer.Terminate(ctx)

    // Get connection string
    redisURL, err := redisContainer.ConnectionString(ctx)
    require.NoError(t, err)

    // Create cache and run tests
    cache := NewRedisCache(redisURL)

    t.Run("Set and Get", func(t *testing.T) {
        err := cache.Set(ctx, "key1", "value1", time.Minute)
        require.NoError(t, err)

        value, err := cache.Get(ctx, "key1")
        require.NoError(t, err)
        require.Equal(t, "value1", value)
    })

    t.Run("Get missing key", func(t *testing.T) {
        _, err := cache.Get(ctx, "missing")
        require.ErrorIs(t, err, ErrCacheMiss)
    })
}

// Full Stack Integration Test
func TestFullStack(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    ctx := context.Background()

    // Start all containers
    pgContainer, _ := postgres.RunContainer(ctx, /* ... */)
    redisContainer, _ := redis.RunContainer(ctx, /* ... */)
    defer pgContainer.Terminate(ctx)
    defer redisContainer.Terminate(ctx)

    // Initialize app with real dependencies
    app := NewApp(Config{
        DatabaseURL: pgConnStr,
        RedisURL:    redisURL,
    })

    server := httptest.NewServer(app.Router())
    defer server.Close()

    // Run full integration tests
    t.Run("Create and Redirect Flow", func(t *testing.T) {
        // ... test implementation
    })
}
```

### Benchmark Tests

Benchmark tests measure performance and help identify bottlenecks.

```go
package shortener

import (
    "testing"
)

func BenchmarkGenerateShortCode(b *testing.B) {
    for i := 0; i < b.N; i++ {
        GenerateShortCode(6)
    }
}

func BenchmarkGenerateShortCode_Parallel(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            GenerateShortCode(6)
        }
    })
}

func BenchmarkValidateURL(b *testing.B) {
    urls := []string{
        "https://example.com",
        "https://example.com/very/long/path/to/resource?with=query&params=here",
        "invalid-url",
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ValidateURL(urls[i%len(urls)])
    }
}

func BenchmarkLinkService_Resolve(b *testing.B) {
    service := NewLinkService(newBenchmarkRepo())

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.Resolve("abc123")
    }
}

// Memory allocation benchmark
func BenchmarkLinkService_Create(b *testing.B) {
    service := NewLinkService(newBenchmarkRepo())

    b.ReportAllocs()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        service.Create("https://example.com")
    }
}
```

Run benchmarks:

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkGenerateShortCode -benchmem ./pkg/shortener

# Compare benchmarks
go test -bench=. -count=5 ./... | tee old.txt
# Make changes
go test -bench=. -count=5 ./... | tee new.txt
benchstat old.txt new.txt
```

---

## Frontend Testing

### Unit Testing with Vitest

[Vitest](https://vitest.dev/) is a fast unit testing framework for Vite projects.

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: ['node_modules/', 'src/test/'],
    },
  },
});
```

```typescript
// src/test/setup.ts
import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

afterEach(() => {
  cleanup();
});
```

```typescript
// src/utils/url.test.ts
import { describe, it, expect } from 'vitest';
import { isValidUrl, formatShortUrl, extractDomain } from './url';

describe('URL Utilities', () => {
  describe('isValidUrl', () => {
    it('should return true for valid HTTP URLs', () => {
      expect(isValidUrl('http://example.com')).toBe(true);
      expect(isValidUrl('https://example.com')).toBe(true);
    });

    it('should return false for invalid URLs', () => {
      expect(isValidUrl('')).toBe(false);
      expect(isValidUrl('not-a-url')).toBe(false);
      expect(isValidUrl('javascript:alert(1)')).toBe(false);
    });
  });

  describe('formatShortUrl', () => {
    it('should format short code to full URL', () => {
      expect(formatShortUrl('abc123')).toBe('https://lnkr.ft/abc123');
    });

    it('should handle custom domains', () => {
      expect(formatShortUrl('abc123', 'custom.com')).toBe('https://custom.com/abc123');
    });
  });

  describe('extractDomain', () => {
    it('should extract domain from URL', () => {
      expect(extractDomain('https://www.example.com/path')).toBe('example.com');
      expect(extractDomain('https://sub.domain.com')).toBe('sub.domain.com');
    });
  });
});
```

```typescript
// src/hooks/useLinks.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useLinks, useCreateLink } from './useLinks';

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });
  return ({ children }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('useLinks', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('should fetch links successfully', async () => {
    const mockLinks = [
      { id: '1', shortCode: 'abc123', originalUrl: 'https://example.com' },
    ];

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockLinks),
    });

    const { result } = renderHook(() => useLinks(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockLinks);
  });

  it('should handle fetch error', async () => {
    global.fetch = vi.fn().mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => useLinks(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
  });
});

describe('useCreateLink', () => {
  it('should create a new link', async () => {
    const newLink = { shortCode: 'xyz789', originalUrl: 'https://test.com' };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(newLink),
    });

    const { result } = renderHook(() => useCreateLink(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ url: 'https://test.com' });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(newLink);
  });
});
```

### Component Testing with React Testing Library

[React Testing Library](https://testing-library.com/docs/react-testing-library/intro/) focuses on testing components as users would interact with them.

```typescript
// src/components/LinkForm.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { LinkForm } from './LinkForm';

describe('LinkForm', () => {
  it('should render the form correctly', () => {
    render(<LinkForm onSubmit={vi.fn()} />);

    expect(screen.getByLabelText(/enter url/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /shorten/i })).toBeInTheDocument();
  });

  it('should submit valid URL', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<LinkForm onSubmit={onSubmit} />);

    const input = screen.getByLabelText(/enter url/i);
    const button = screen.getByRole('button', { name: /shorten/i });

    await user.type(input, 'https://example.com');
    await user.click(button);

    expect(onSubmit).toHaveBeenCalledWith({ url: 'https://example.com' });
  });

  it('should show validation error for invalid URL', async () => {
    const user = userEvent.setup();
    render(<LinkForm onSubmit={vi.fn()} />);

    const input = screen.getByLabelText(/enter url/i);
    const button = screen.getByRole('button', { name: /shorten/i });

    await user.type(input, 'not-a-url');
    await user.click(button);

    expect(screen.getByText(/please enter a valid url/i)).toBeInTheDocument();
  });

  it('should disable button while submitting', async () => {
    render(<LinkForm onSubmit={vi.fn()} isLoading={true} />);

    const button = screen.getByRole('button', { name: /shortening/i });
    expect(button).toBeDisabled();
  });
});

// src/components/LinkList.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { LinkList } from './LinkList';

const mockLinks = [
  {
    id: '1',
    shortCode: 'abc123',
    originalUrl: 'https://example.com/very/long/url',
    clicks: 42,
    createdAt: '2025-01-15T10:00:00Z',
  },
  {
    id: '2',
    shortCode: 'xyz789',
    originalUrl: 'https://test.com',
    clicks: 10,
    createdAt: '2025-01-16T10:00:00Z',
  },
];

describe('LinkList', () => {
  it('should render empty state when no links', () => {
    render(<LinkList links={[]} />);
    expect(screen.getByText(/no links yet/i)).toBeInTheDocument();
  });

  it('should render list of links', () => {
    render(<LinkList links={mockLinks} />);

    expect(screen.getByText('abc123')).toBeInTheDocument();
    expect(screen.getByText('xyz789')).toBeInTheDocument();
    expect(screen.getByText('42 clicks')).toBeInTheDocument();
  });

  it('should copy short URL to clipboard', async () => {
    const user = userEvent.setup();
    const mockClipboard = vi.fn();
    Object.assign(navigator, {
      clipboard: { writeText: mockClipboard },
    });

    render(<LinkList links={mockLinks} />);

    const firstRow = screen.getAllByRole('row')[1];
    const copyButton = within(firstRow).getByRole('button', { name: /copy/i });

    await user.click(copyButton);

    expect(mockClipboard).toHaveBeenCalledWith('https://lnkr.ft/abc123');
  });

  it('should call onDelete when delete button clicked', async () => {
    const user = userEvent.setup();
    const onDelete = vi.fn();
    render(<LinkList links={mockLinks} onDelete={onDelete} />);

    const firstRow = screen.getAllByRole('row')[1];
    const deleteButton = within(firstRow).getByRole('button', { name: /delete/i });

    await user.click(deleteButton);

    expect(onDelete).toHaveBeenCalledWith('1');
  });
});

// src/components/Analytics.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { AnalyticsDashboard } from './AnalyticsDashboard';

describe('AnalyticsDashboard', () => {
  const mockAnalytics = {
    totalClicks: 1250,
    uniqueVisitors: 890,
    topLinks: [
      { shortCode: 'abc123', clicks: 500 },
      { shortCode: 'xyz789', clicks: 300 },
    ],
    clicksByCountry: [
      { country: 'US', clicks: 600 },
      { country: 'UK', clicks: 200 },
    ],
    clicksByDevice: {
      desktop: 700,
      mobile: 450,
      tablet: 100,
    },
  };

  it('should display total clicks', () => {
    render(<AnalyticsDashboard data={mockAnalytics} />);
    expect(screen.getByText('1,250')).toBeInTheDocument();
    expect(screen.getByText(/total clicks/i)).toBeInTheDocument();
  });

  it('should display top performing links', () => {
    render(<AnalyticsDashboard data={mockAnalytics} />);
    expect(screen.getByText('abc123')).toBeInTheDocument();
    expect(screen.getByText('500 clicks')).toBeInTheDocument();
  });

  it('should render charts', () => {
    render(<AnalyticsDashboard data={mockAnalytics} />);
    expect(screen.getByTestId('clicks-chart')).toBeInTheDocument();
    expect(screen.getByTestId('geo-chart')).toBeInTheDocument();
  });
});
```

### End-to-End Testing with Playwright

[Playwright](https://playwright.dev/) provides reliable end-to-end testing across browsers.

```typescript
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
    },
  ],
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: !process.env.CI,
  },
});
```

```typescript
// e2e/link-creation.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Link Creation Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should create a short link', async ({ page }) => {
    // Enter URL
    await page.getByLabel(/enter url/i).fill('https://example.com/test');

    // Click shorten button
    await page.getByRole('button', { name: /shorten/i }).click();

    // Wait for result
    await expect(page.getByText(/link created/i)).toBeVisible();

    // Verify short URL is displayed
    const shortUrl = page.getByTestId('short-url');
    await expect(shortUrl).toContainText('lnkr.ft/');

    // Copy button should work
    await page.getByRole('button', { name: /copy/i }).click();
    await expect(page.getByText(/copied/i)).toBeVisible();
  });

  test('should show error for invalid URL', async ({ page }) => {
    await page.getByLabel(/enter url/i).fill('not-a-valid-url');
    await page.getByRole('button', { name: /shorten/i }).click();

    await expect(page.getByText(/please enter a valid url/i)).toBeVisible();
  });

  test('should allow custom short code', async ({ page }) => {
    await page.getByLabel(/enter url/i).fill('https://example.com');

    // Expand advanced options
    await page.getByRole('button', { name: /advanced/i }).click();

    // Enter custom code
    await page.getByLabel(/custom code/i).fill('my-custom-code');

    await page.getByRole('button', { name: /shorten/i }).click();

    await expect(page.getByTestId('short-url')).toContainText('my-custom-code');
  });
});

// e2e/analytics.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Analytics Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.getByLabel(/email/i).fill('test@example.com');
    await page.getByLabel(/password/i).fill('password123');
    await page.getByRole('button', { name: /sign in/i }).click();

    await expect(page).toHaveURL('/dashboard');
  });

  test('should display analytics overview', async ({ page }) => {
    await page.getByRole('link', { name: /analytics/i }).click();

    await expect(page.getByText(/total clicks/i)).toBeVisible();
    await expect(page.getByText(/unique visitors/i)).toBeVisible();
    await expect(page.getByTestId('clicks-chart')).toBeVisible();
  });

  test('should filter by date range', async ({ page }) => {
    await page.getByRole('link', { name: /analytics/i }).click();

    // Select date range
    await page.getByRole('combobox', { name: /date range/i }).click();
    await page.getByRole('option', { name: /last 30 days/i }).click();

    // Chart should update
    await expect(page.getByTestId('clicks-chart')).toBeVisible();
  });

  test('should export data as CSV', async ({ page }) => {
    await page.getByRole('link', { name: /analytics/i }).click();

    const downloadPromise = page.waitForEvent('download');
    await page.getByRole('button', { name: /export csv/i }).click();

    const download = await downloadPromise;
    expect(download.suggestedFilename()).toMatch(/analytics.*\.csv/);
  });
});

// e2e/redirect.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Link Redirection', () => {
  test('should redirect short URL to original', async ({ page, request }) => {
    // Create a link first via API
    const response = await request.post('/api/v1/links', {
      data: { url: 'https://example.com/redirect-test' },
    });
    const { short_code } = await response.json();

    // Navigate to short URL
    await page.goto(`/${short_code}`);

    // Should redirect to original URL
    await expect(page).toHaveURL('https://example.com/redirect-test');
  });

  test('should show 404 for non-existent short code', async ({ page }) => {
    await page.goto('/nonexistent123');

    await expect(page.getByText(/link not found/i)).toBeVisible();
    await expect(page).toHaveURL('/404');
  });
});
```

Run Playwright tests:

```bash
# Install browsers
npx playwright install

# Run all tests
npx playwright test

# Run with UI mode
npx playwright test --ui

# Run specific test file
npx playwright test e2e/link-creation.spec.ts

# Run in headed mode
npx playwright test --headed

# Generate report
npx playwright show-report
```

---

## Load Testing

### k6 Load Testing

[k6](https://k6.io/) is a modern load testing tool with JavaScript scripting.

```javascript
// load-tests/smoke.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 1,
  duration: '30s',
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  // Test link creation
  const createPayload = JSON.stringify({
    url: 'https://example.com/test-' + Date.now(),
  });

  const createResponse = http.post(
    'http://localhost:8080/api/v1/links',
    createPayload,
    { headers: { 'Content-Type': 'application/json' } }
  );

  check(createResponse, {
    'create status is 201': (r) => r.status === 201,
    'create response has short_code': (r) => JSON.parse(r.body).short_code !== undefined,
  });

  // Test redirect
  if (createResponse.status === 201) {
    const { short_code } = JSON.parse(createResponse.body);
    const redirectResponse = http.get(`http://localhost:8080/${short_code}`, {
      redirects: 0,
    });

    check(redirectResponse, {
      'redirect status is 301': (r) => r.status === 301,
      'redirect has location header': (r) => r.headers['Location'] !== undefined,
    });
  }

  sleep(1);
}
```

```javascript
// load-tests/stress.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const redirectDuration = new Trend('redirect_duration');

export const options = {
  stages: [
    { duration: '2m', target: 100 },   // Ramp up to 100 users
    { duration: '5m', target: 100 },   // Stay at 100 users
    { duration: '2m', target: 200 },   // Ramp up to 200 users
    { duration: '5m', target: 200 },   // Stay at 200 users
    { duration: '2m', target: 0 },     // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<200', 'p(99)<500'],
    errors: ['rate<0.1'],
    redirect_duration: ['p(95)<50'],
  },
};

const BASE_URL = 'http://localhost:8080';
const SHORT_CODES = ['abc123', 'xyz789', 'test01']; // Pre-created codes

export default function () {
  // Simulate realistic traffic: 90% redirects, 10% creates
  if (Math.random() < 0.9) {
    // Redirect request
    const code = SHORT_CODES[Math.floor(Math.random() * SHORT_CODES.length)];
    const start = Date.now();

    const response = http.get(`${BASE_URL}/${code}`, { redirects: 0 });

    redirectDuration.add(Date.now() - start);

    const success = check(response, {
      'redirect succeeded': (r) => r.status === 301 || r.status === 302,
    });

    errorRate.add(!success);
  } else {
    // Create request
    const payload = JSON.stringify({
      url: `https://example.com/stress-${Date.now()}-${__VU}`,
    });

    const response = http.post(`${BASE_URL}/api/v1/links`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });

    const success = check(response, {
      'create succeeded': (r) => r.status === 201,
    });

    errorRate.add(!success);
  }

  sleep(Math.random() * 2);
}
```

```javascript
// load-tests/soak.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '5m', target: 50 },    // Ramp up
    { duration: '4h', target: 50 },    // Soak for 4 hours
    { duration: '5m', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  const response = http.get('http://localhost:8080/abc123', { redirects: 0 });

  check(response, {
    'status is 301': (r) => r.status === 301,
  });

  sleep(1);
}
```

Run k6 tests:

```bash
# Run smoke test
k6 run load-tests/smoke.js

# Run stress test with output
k6 run --out json=results.json load-tests/stress.js

# Run with cloud integration
k6 cloud load-tests/stress.js
```

### Vegeta Load Testing

[Vegeta](https://github.com/tsenart/vegeta) is a versatile HTTP load testing tool.

```bash
# Install vegeta
go install github.com/tsenart/vegeta@latest

# Simple attack
echo "GET http://localhost:8080/abc123" | vegeta attack -duration=30s -rate=100 | vegeta report

# Attack with custom headers
cat << EOF | vegeta attack -duration=1m -rate=50 | vegeta report
POST http://localhost:8080/api/v1/links
Content-Type: application/json
@payload.json
EOF

# Generate HTML report
echo "GET http://localhost:8080/abc123" | vegeta attack -duration=30s -rate=100 | vegeta encode | vegeta plot > results.html

# Multiple targets
cat << EOF > targets.txt
GET http://localhost:8080/abc123
GET http://localhost:8080/xyz789
GET http://localhost:8080/test01
POST http://localhost:8080/api/v1/links
Content-Type: application/json
@create-payload.json
EOF

vegeta attack -targets=targets.txt -duration=2m -rate=200 | vegeta report
```

```go
// Programmatic vegeta usage in Go tests
package loadtest

import (
    "fmt"
    "testing"
    "time"

    vegeta "github.com/tsenart/vegeta/v12/lib"
)

func TestLoadRedirect(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping load test in short mode")
    }

    rate := vegeta.Rate{Freq: 100, Per: time.Second}
    duration := 30 * time.Second

    targeter := vegeta.NewStaticTargeter(vegeta.Target{
        Method: "GET",
        URL:    "http://localhost:8080/abc123",
    })

    attacker := vegeta.NewAttacker()

    var metrics vegeta.Metrics
    for res := range attacker.Attack(targeter, rate, duration, "Redirect Load Test") {
        metrics.Add(res)
    }
    metrics.Close()

    // Assert on metrics
    if metrics.Latencies.P95 > 100*time.Millisecond {
        t.Errorf("P95 latency too high: %v", metrics.Latencies.P95)
    }

    if metrics.Success < 0.99 {
        t.Errorf("Success rate too low: %.2f%%", metrics.Success*100)
    }

    fmt.Printf("Requests: %d\n", metrics.Requests)
    fmt.Printf("Duration: %v\n", metrics.Duration)
    fmt.Printf("Latencies P50: %v\n", metrics.Latencies.P50)
    fmt.Printf("Latencies P95: %v\n", metrics.Latencies.P95)
    fmt.Printf("Latencies P99: %v\n", metrics.Latencies.P99)
    fmt.Printf("Success Rate: %.2f%%\n", metrics.Success*100)
}
```

---

## Coverage Reporting

### Go Coverage

```bash
# Run tests with coverage
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Coverage for specific packages
go test -coverprofile=coverage.out -coverpkg=./pkg/... ./...

# Race detection with coverage
go test -race -coverprofile=coverage.out ./...
```

```makefile
# Makefile
.PHONY: test coverage

test:
	go test -v -race ./...

coverage:
	go test -coverprofile=coverage.out -coverpkg=./... ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'
```

### Frontend Coverage

```bash
# Run tests with coverage
npm run test -- --coverage

# Coverage report location
# - coverage/lcov-report/index.html (HTML)
# - coverage/lcov.info (LCOV format)
```

### CI Coverage Integration

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test-backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run tests with coverage
        run: go test -coverprofile=coverage.out -covermode=atomic ./...

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          files: coverage.out
          flags: backend

  test-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'

      - run: npm ci
      - run: npm run test -- --coverage

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          files: coverage/lcov.info
          flags: frontend
```

---

## Continuous Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  GO_VERSION: '1.22'
  NODE_VERSION: '20'

jobs:
  lint-backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4

  lint-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
      - run: npm ci
      - run: npm run lint
      - run: npm run type-check

  test-backend:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
      redis:
        image: redis:7
        ports:
          - 6379:6379
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      - run: go test -race -coverprofile=coverage.out ./...
      - uses: codecov/codecov-action@v4
        with:
          files: coverage.out

  test-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
      - run: npm ci
      - run: npm run test -- --coverage
      - uses: codecov/codecov-action@v4
        with:
          files: coverage/lcov.info

  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
      - run: npm ci
      - run: npx playwright install --with-deps
      - run: npm run build
      - run: npm run test:e2e
      - uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: playwright-report
          path: playwright-report/

  benchmark:
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      - name: Run benchmarks
        run: go test -bench=. -benchmem ./... | tee benchmark.txt
      - uses: benchmark-action/github-action-benchmark@v1
        with:
          tool: 'go'
          output-file-path: benchmark.txt
          github-token: ${{ secrets.GITHUB_TOKEN }}
          auto-push: true
```

---

## Summary

This testing strategy ensures Linkrift maintains high quality through:

- **Unit Tests**: Fast, isolated tests for business logic
- **Integration Tests**: Verify components work together with real dependencies
- **E2E Tests**: Validate complete user workflows
- **Load Tests**: Ensure performance under realistic conditions
- **Continuous Coverage**: Track and improve test coverage over time

For questions or suggestions, open an issue or contribute improvements via pull request.
