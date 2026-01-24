# Frontend Architecture

> Last Updated: 2025-01-24

React frontend architecture, patterns, and best practices for Linkrift.

## Table of Contents

- [Project Structure](#project-structure)
- [Routing Architecture](#routing-architecture)
- [State Management](#state-management)
- [API Client](#api-client)
- [Component Patterns](#component-patterns)
- [Form Handling](#form-handling)
- [Error Handling](#error-handling)
- [Performance Patterns](#performance-patterns)

---

## Project Structure

```
web/src/
├── components/
│   ├── ui/                    # Shadcn UI components
│   │   ├── button.tsx
│   │   ├── input.tsx
│   │   ├── dialog.tsx
│   │   └── ...
│   ├── forms/                 # Form components
│   │   ├── LinkForm.tsx
│   │   ├── DomainForm.tsx
│   │   └── ...
│   ├── layouts/               # Layout components
│   │   ├── DashboardLayout.tsx
│   │   ├── AuthLayout.tsx
│   │   └── PublicLayout.tsx
│   └── features/              # Feature-specific components
│       ├── links/
│       │   ├── LinkCard.tsx
│       │   ├── LinkList.tsx
│       │   ├── LinkFilters.tsx
│       │   └── CreateLinkDialog.tsx
│       ├── analytics/
│       │   ├── AnalyticsChart.tsx
│       │   ├── StatsCard.tsx
│       │   └── GeoMap.tsx
│       ├── domains/
│       ├── qr-codes/
│       └── settings/
│
├── pages/                     # Page components (routes)
│   ├── auth/
│   │   ├── Login.tsx
│   │   ├── Register.tsx
│   │   └── ForgotPassword.tsx
│   ├── dashboard/
│   │   ├── Dashboard.tsx
│   │   ├── Links.tsx
│   │   ├── LinkDetail.tsx
│   │   ├── Analytics.tsx
│   │   ├── Domains.tsx
│   │   ├── QRCodes.tsx
│   │   └── Settings.tsx
│   └── public/
│       ├── Home.tsx
│       └── Pricing.tsx
│
├── hooks/                     # Custom React hooks
│   ├── useAuth.ts
│   ├── useLicense.ts
│   ├── useLinks.ts
│   ├── useAnalytics.ts
│   ├── useDebounce.ts
│   └── useMediaQuery.ts
│
├── stores/                    # Zustand stores
│   ├── authStore.ts
│   ├── licenseStore.ts
│   ├── uiStore.ts
│   └── workspaceStore.ts
│
├── services/                  # API clients
│   ├── api.ts                 # Base API client
│   ├── auth.ts
│   ├── links.ts
│   ├── analytics.ts
│   └── domains.ts
│
├── lib/                       # Utilities
│   ├── utils.ts               # cn(), formatDate(), etc.
│   ├── constants.ts
│   └── license.ts
│
├── types/                     # TypeScript types
│   ├── api.ts
│   ├── link.ts
│   ├── user.ts
│   └── analytics.ts
│
├── styles/
│   └── globals.css
│
├── App.tsx
├── main.tsx
└── router.tsx
```

---

## Routing Architecture

### Route Configuration

```typescript
// router.tsx
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import { lazy, Suspense } from 'react';

// Lazy load pages
const Dashboard = lazy(() => import('./pages/dashboard/Dashboard'));
const Links = lazy(() => import('./pages/dashboard/Links'));
const Analytics = lazy(() => import('./pages/dashboard/Analytics'));
const Settings = lazy(() => import('./pages/dashboard/Settings'));

// Layouts
import DashboardLayout from './components/layouts/DashboardLayout';
import AuthLayout from './components/layouts/AuthLayout';

// Guards
import { ProtectedRoute } from './components/auth/ProtectedRoute';
import { PublicOnlyRoute } from './components/auth/PublicOnlyRoute';

const router = createBrowserRouter([
  // Public routes
  {
    path: '/',
    element: <PublicLayout />,
    children: [
      { index: true, element: <Home /> },
      { path: 'pricing', element: <Pricing /> },
    ],
  },

  // Auth routes
  {
    path: '/auth',
    element: <PublicOnlyRoute><AuthLayout /></PublicOnlyRoute>,
    children: [
      { path: 'login', element: <Login /> },
      { path: 'register', element: <Register /> },
      { path: 'forgot-password', element: <ForgotPassword /> },
    ],
  },

  // Dashboard routes (protected)
  {
    path: '/dashboard',
    element: <ProtectedRoute><DashboardLayout /></ProtectedRoute>,
    children: [
      { index: true, element: <Dashboard /> },
      { path: 'links', element: <Links /> },
      { path: 'links/:id', element: <LinkDetail /> },
      { path: 'analytics', element: <Analytics /> },
      { path: 'domains', element: <Domains /> },
      { path: 'qr-codes', element: <QRCodes /> },
      { path: 'settings/*', element: <Settings /> },
    ],
  },

  // 404
  { path: '*', element: <NotFound /> },
]);

export function Router() {
  return (
    <Suspense fallback={<PageLoader />}>
      <RouterProvider router={router} />
    </Suspense>
  );
}
```

### Protected Route Component

```typescript
// components/auth/ProtectedRoute.tsx
import { Navigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredRole?: string;
}

export function ProtectedRoute({ children, requiredRole }: ProtectedRouteProps) {
  const { isAuthenticated, user, isLoading } = useAuthStore();
  const location = useLocation();

  if (isLoading) {
    return <PageLoader />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/auth/login" state={{ from: location }} replace />;
  }

  if (requiredRole && user?.role !== requiredRole) {
    return <Navigate to="/dashboard" replace />;
  }

  return <>{children}</>;
}
```

---

## State Management

### Auth Store (Zustand)

```typescript
// stores/authStore.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { api } from '@/services/api';
import type { User } from '@/types/user';

interface AuthState {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  // Actions
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<void>;
  updateUser: (user: Partial<User>) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      isAuthenticated: false,
      isLoading: true,

      login: async (email: string, password: string) => {
        const response = await api.post('/auth/login', { email, password });
        set({
          user: response.user,
          accessToken: response.access_token,
          isAuthenticated: true,
        });
      },

      logout: async () => {
        try {
          await api.post('/auth/logout');
        } finally {
          set({
            user: null,
            accessToken: null,
            isAuthenticated: false,
          });
        }
      },

      refreshToken: async () => {
        try {
          const response = await api.post('/auth/refresh');
          set({ accessToken: response.access_token });
        } catch {
          get().logout();
        }
      },

      updateUser: (updates) => {
        set((state) => ({
          user: state.user ? { ...state.user, ...updates } : null,
        }));
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
```

### UI Store

```typescript
// stores/uiStore.ts
import { create } from 'zustand';

interface UIState {
  sidebarOpen: boolean;
  theme: 'light' | 'dark' | 'system';
  commandPaletteOpen: boolean;

  toggleSidebar: () => void;
  setTheme: (theme: 'light' | 'dark' | 'system') => void;
  toggleCommandPalette: () => void;
}

export const useUIStore = create<UIState>((set) => ({
  sidebarOpen: true,
  theme: 'system',
  commandPaletteOpen: false,

  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
  setTheme: (theme) => set({ theme }),
  toggleCommandPalette: () => set((state) => ({
    commandPaletteOpen: !state.commandPaletteOpen
  })),
}));
```

### Server State (TanStack Query)

```typescript
// hooks/useLinks.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { linksApi } from '@/services/links';
import type { Link, CreateLinkInput, LinkFilters } from '@/types/link';

// Query keys factory
export const linkKeys = {
  all: ['links'] as const,
  lists: () => [...linkKeys.all, 'list'] as const,
  list: (filters: LinkFilters) => [...linkKeys.lists(), filters] as const,
  details: () => [...linkKeys.all, 'detail'] as const,
  detail: (id: string) => [...linkKeys.details(), id] as const,
};

// List hook
export function useLinks(filters: LinkFilters) {
  return useQuery({
    queryKey: linkKeys.list(filters),
    queryFn: () => linksApi.list(filters),
    staleTime: 30_000, // 30 seconds
  });
}

// Detail hook
export function useLink(id: string) {
  return useQuery({
    queryKey: linkKeys.detail(id),
    queryFn: () => linksApi.get(id),
    enabled: !!id,
  });
}

// Create mutation
export function useCreateLink() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateLinkInput) => linksApi.create(input),
    onSuccess: () => {
      // Invalidate list queries
      queryClient.invalidateQueries({ queryKey: linkKeys.lists() });
    },
  });
}

// Update mutation with optimistic update
export function useUpdateLink() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, ...input }: UpdateLinkInput & { id: string }) =>
      linksApi.update(id, input),

    onMutate: async ({ id, ...input }) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: linkKeys.detail(id) });

      // Snapshot previous value
      const previousLink = queryClient.getQueryData<Link>(linkKeys.detail(id));

      // Optimistically update
      queryClient.setQueryData<Link>(linkKeys.detail(id), (old) =>
        old ? { ...old, ...input } : old
      );

      return { previousLink };
    },

    onError: (err, { id }, context) => {
      // Rollback on error
      if (context?.previousLink) {
        queryClient.setQueryData(linkKeys.detail(id), context.previousLink);
      }
    },

    onSettled: (_, __, { id }) => {
      // Refetch after mutation
      queryClient.invalidateQueries({ queryKey: linkKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: linkKeys.lists() });
    },
  });
}

// Delete mutation
export function useDeleteLink() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => linksApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: linkKeys.lists() });
    },
  });
}
```

---

## API Client

### Base API Client

```typescript
// services/api.ts
import { useAuthStore } from '@/stores/authStore';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

interface RequestOptions extends RequestInit {
  params?: Record<string, string>;
}

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options: RequestOptions = {}
  ): Promise<T> {
    const { params, ...fetchOptions } = options;

    // Build URL with query params
    let url = `${this.baseUrl}/api/v1${endpoint}`;
    if (params) {
      const searchParams = new URLSearchParams(params);
      url += `?${searchParams.toString()}`;
    }

    // Get auth token
    const token = useAuthStore.getState().accessToken;

    // Make request
    const response = await fetch(url, {
      ...fetchOptions,
      headers: {
        'Content-Type': 'application/json',
        ...(token && { Authorization: `Bearer ${token}` }),
        ...fetchOptions.headers,
      },
    });

    // Handle 401 - token expired
    if (response.status === 401) {
      try {
        await useAuthStore.getState().refreshToken();
        // Retry request with new token
        return this.request<T>(endpoint, options);
      } catch {
        useAuthStore.getState().logout();
        throw new Error('Session expired');
      }
    }

    // Handle errors
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new ApiError(response.status, error.message || 'Request failed', error);
    }

    // Parse response
    if (response.status === 204) {
      return undefined as T;
    }

    return response.json();
  }

  get<T>(endpoint: string, params?: Record<string, string>): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET', params });
  }

  post<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  put<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  patch<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  }
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public data?: unknown
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export const api = new ApiClient(API_BASE_URL);
```

### Service Modules

```typescript
// services/links.ts
import { api } from './api';
import type { Link, CreateLinkInput, UpdateLinkInput, LinkFilters, PaginatedResponse } from '@/types/link';

export const linksApi = {
  list: (filters: LinkFilters): Promise<PaginatedResponse<Link>> =>
    api.get('/links', {
      page: String(filters.page || 1),
      limit: String(filters.limit || 20),
      ...(filters.search && { search: filters.search }),
      ...(filters.domain && { domain: filters.domain }),
    }),

  get: (id: string): Promise<Link> =>
    api.get(`/links/${id}`),

  create: (input: CreateLinkInput): Promise<Link> =>
    api.post('/links', input),

  update: (id: string, input: UpdateLinkInput): Promise<Link> =>
    api.patch(`/links/${id}`, input),

  delete: (id: string): Promise<void> =>
    api.delete(`/links/${id}`),

  bulk: (inputs: CreateLinkInput[]): Promise<Link[]> =>
    api.post('/links/bulk', { links: inputs }),
};
```

---

## Component Patterns

### Feature Component

```typescript
// components/features/links/LinkCard.tsx
import { Link } from '@/types/link';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Copy, ExternalLink, MoreVertical, QrCode } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import { useToast } from '@/hooks/useToast';
import { cn } from '@/lib/utils';

interface LinkCardProps {
  link: Link;
  onEdit?: (link: Link) => void;
  onDelete?: (link: Link) => void;
  className?: string;
}

export function LinkCard({ link, onEdit, onDelete, className }: LinkCardProps) {
  const { toast } = useToast();

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(link.short_url);
    toast({ title: 'Copied to clipboard' });
  };

  return (
    <Card className={cn('hover:shadow-md transition-shadow', className)}>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <div className="flex items-center gap-3">
          {link.favicon_url && (
            <img
              src={link.favicon_url}
              alt=""
              className="h-8 w-8 rounded"
            />
          )}
          <div>
            <h3 className="font-medium text-sm truncate max-w-[200px]">
              {link.title || link.url}
            </h3>
            <a
              href={link.short_url}
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-primary hover:underline"
            >
              {link.short_url}
            </a>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="ghost" size="icon" onClick={copyToClipboard}>
            <Copy className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" asChild>
            <a href={link.short_url} target="_blank" rel="noopener noreferrer">
              <ExternalLink className="h-4 w-4" />
            </a>
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onEdit?.(link)}>
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem>
                <QrCode className="h-4 w-4 mr-2" />
                QR Code
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="text-destructive"
                onClick={() => onDelete?.(link)}
              >
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>

      <CardContent>
        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <span>{link.total_clicks.toLocaleString()} clicks</span>
          <span>Created {formatDistanceToNow(new Date(link.created_at))} ago</span>
        </div>
      </CardContent>
    </Card>
  );
}
```

### Compound Component Pattern

```typescript
// components/ui/data-table.tsx
import { createContext, useContext, useState } from 'react';

interface DataTableContextValue<T> {
  data: T[];
  selectedRows: Set<string>;
  toggleRow: (id: string) => void;
  toggleAll: () => void;
  isAllSelected: boolean;
}

const DataTableContext = createContext<DataTableContextValue<any> | null>(null);

function useDataTable<T>() {
  const context = useContext(DataTableContext);
  if (!context) {
    throw new Error('useDataTable must be used within DataTable');
  }
  return context as DataTableContextValue<T>;
}

interface DataTableProps<T> {
  data: T[];
  getRowId: (row: T) => string;
  children: React.ReactNode;
}

export function DataTable<T>({ data, getRowId, children }: DataTableProps<T>) {
  const [selectedRows, setSelectedRows] = useState<Set<string>>(new Set());

  const toggleRow = (id: string) => {
    setSelectedRows((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const toggleAll = () => {
    if (selectedRows.size === data.length) {
      setSelectedRows(new Set());
    } else {
      setSelectedRows(new Set(data.map(getRowId)));
    }
  };

  const isAllSelected = selectedRows.size === data.length && data.length > 0;

  return (
    <DataTableContext.Provider
      value={{ data, selectedRows, toggleRow, toggleAll, isAllSelected }}
    >
      <div className="rounded-md border">{children}</div>
    </DataTableContext.Provider>
  );
}

DataTable.Header = function DataTableHeader({ children }: { children: React.ReactNode }) {
  return <div className="border-b bg-muted/50 px-4 py-3">{children}</div>;
};

DataTable.Body = function DataTableBody<T>({
  renderRow,
}: {
  renderRow: (row: T, index: number) => React.ReactNode;
}) {
  const { data } = useDataTable<T>();
  return <div>{data.map((row, index) => renderRow(row, index))}</div>;
};

DataTable.Row = function DataTableRow({
  id,
  children,
}: {
  id: string;
  children: React.ReactNode;
}) {
  const { selectedRows } = useDataTable();
  const isSelected = selectedRows.has(id);

  return (
    <div
      className={cn(
        'flex items-center px-4 py-3 border-b last:border-0',
        isSelected && 'bg-primary/5'
      )}
    >
      {children}
    </div>
  );
};
```

---

## Form Handling

### React Hook Form + Zod

```typescript
// components/forms/LinkForm.tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { useCreateLink } from '@/hooks/useLinks';

const linkSchema = z.object({
  url: z.string().url('Please enter a valid URL'),
  customCode: z
    .string()
    .min(3, 'Must be at least 3 characters')
    .max(50, 'Must be at most 50 characters')
    .regex(/^[a-zA-Z0-9-_]+$/, 'Only letters, numbers, dashes, and underscores')
    .optional()
    .or(z.literal('')),
  title: z.string().max(500).optional(),
  expiresAt: z.date().optional(),
});

type LinkFormValues = z.infer<typeof linkSchema>;

interface LinkFormProps {
  onSuccess?: () => void;
}

export function LinkForm({ onSuccess }: LinkFormProps) {
  const createLink = useCreateLink();

  const form = useForm<LinkFormValues>({
    resolver: zodResolver(linkSchema),
    defaultValues: {
      url: '',
      customCode: '',
      title: '',
    },
  });

  const onSubmit = async (values: LinkFormValues) => {
    try {
      await createLink.mutateAsync({
        url: values.url,
        custom_code: values.customCode || undefined,
        title: values.title || undefined,
        expires_at: values.expiresAt?.toISOString(),
      });
      form.reset();
      onSuccess?.();
    } catch (error) {
      // Error is handled by mutation
    }
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="url"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Destination URL</FormLabel>
              <FormControl>
                <Input placeholder="https://example.com" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="customCode"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Custom Short Code (optional)</FormLabel>
              <FormControl>
                <Input placeholder="my-custom-link" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="title"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Title (optional)</FormLabel>
              <FormControl>
                <Input placeholder="My awesome link" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button type="submit" disabled={createLink.isPending}>
          {createLink.isPending ? 'Creating...' : 'Create Link'}
        </Button>
      </form>
    </Form>
  );
}
```

---

## Error Handling

### Error Boundary

```typescript
// components/ErrorBoundary.tsx
import { Component, ErrorInfo, ReactNode } from 'react';
import { Button } from '@/components/ui/button';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
    // Report to Sentry
    // Sentry.captureException(error, { extra: errorInfo });
  }

  render() {
    if (this.state.hasError) {
      return (
        this.props.fallback || (
          <div className="flex flex-col items-center justify-center min-h-[400px] p-8">
            <h2 className="text-xl font-semibold mb-2">Something went wrong</h2>
            <p className="text-muted-foreground mb-4">
              {this.state.error?.message || 'An unexpected error occurred'}
            </p>
            <Button onClick={() => this.setState({ hasError: false })}>
              Try again
            </Button>
          </div>
        )
      );
    }

    return this.props.children;
  }
}
```

### Query Error Handling

```typescript
// hooks/useQueryErrorHandler.ts
import { useEffect } from 'react';
import { useToast } from '@/hooks/useToast';

export function useQueryErrorHandler(error: Error | null) {
  const { toast } = useToast();

  useEffect(() => {
    if (error) {
      toast({
        variant: 'destructive',
        title: 'Error',
        description: error.message || 'Something went wrong',
      });
    }
  }, [error, toast]);
}
```

---

## Performance Patterns

### Code Splitting

```typescript
// Lazy load heavy components
const AnalyticsChart = lazy(() => import('./AnalyticsChart'));
const QRCodeEditor = lazy(() => import('./QRCodeEditor'));

// Use with Suspense
<Suspense fallback={<ChartSkeleton />}>
  <AnalyticsChart data={data} />
</Suspense>
```

### Memoization

```typescript
// Memoize expensive computations
const sortedLinks = useMemo(
  () => links.sort((a, b) => b.created_at.localeCompare(a.created_at)),
  [links]
);

// Memoize callbacks
const handleSearch = useCallback(
  debounce((query: string) => {
    setFilters((prev) => ({ ...prev, search: query }));
  }, 300),
  []
);

// Memoize components
const LinkCard = memo(function LinkCard({ link }: { link: Link }) {
  return <Card>...</Card>;
});
```

### Virtual Lists

```typescript
// For large lists, use react-window
import { FixedSizeList } from 'react-window';

function LinkList({ links }: { links: Link[] }) {
  return (
    <FixedSizeList
      height={600}
      width="100%"
      itemCount={links.length}
      itemSize={80}
    >
      {({ index, style }) => (
        <div style={style}>
          <LinkCard link={links[index]} />
        </div>
      )}
    </FixedSizeList>
  );
}
```

---

## Related Documentation

- [Design System](../design/DESIGN_SYSTEM.md) — Visual design
- [UI Components](../design/UI_COMPONENTS.md) — Component library
- [Testing](../testing/TESTING.md) — Frontend testing
- [Performance Optimization](../reference/PERFORMANCE_OPTIMIZATION.md) — Performance tips
