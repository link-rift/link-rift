# UI Components

> Last Updated: 2025-01-24

Component library specifications and implementation details for Linkrift.

## Table of Contents

- [Core Components](#core-components)
- [Layout Components](#layout-components)
- [Form Components](#form-components)
- [Data Display](#data-display)
- [Feedback Components](#feedback-components)
- [Dashboard Components](#dashboard-components)
- [Loading States](#loading-states)
- [Empty States](#empty-states)

---

## Core Components

### Button

```tsx
// components/ui/button.tsx
import { cva, type VariantProps } from 'class-variance-authority';

const buttonVariants = cva(
  'inline-flex items-center justify-center rounded-md text-sm font-medium transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        default:
          'bg-brand-500 text-white hover:bg-brand-600 focus-visible:ring-brand-500',
        destructive:
          'bg-red-500 text-white hover:bg-red-600 focus-visible:ring-red-500',
        outline:
          'border border-gray-300 bg-white text-gray-700 hover:bg-gray-50 focus-visible:ring-gray-500',
        secondary:
          'bg-gray-100 text-gray-900 hover:bg-gray-200 focus-visible:ring-gray-500',
        ghost:
          'text-gray-600 hover:bg-gray-100 hover:text-gray-900',
        link:
          'text-brand-600 underline-offset-4 hover:underline',
      },
      size: {
        sm: 'h-8 px-3 text-xs',
        default: 'h-10 px-4',
        lg: 'h-12 px-6 text-base',
        icon: 'h-10 w-10',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  }
);

interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  isLoading?: boolean;
}

export function Button({
  className,
  variant,
  size,
  isLoading,
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      className={cn(buttonVariants({ variant, size }), className)}
      disabled={isLoading || props.disabled}
      {...props}
    >
      {isLoading ? (
        <>
          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          Loading...
        </>
      ) : (
        children
      )}
    </button>
  );
}
```

### Badge

```tsx
// components/ui/badge.tsx
const badgeVariants = cva(
  'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors',
  {
    variants: {
      variant: {
        default: 'bg-gray-100 text-gray-800',
        brand: 'bg-brand-100 text-brand-800',
        success: 'bg-green-100 text-green-800',
        warning: 'bg-amber-100 text-amber-800',
        error: 'bg-red-100 text-red-800',
        outline: 'border border-gray-300 text-gray-700',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
);

export function Badge({ className, variant, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ variant }), className)} {...props} />;
}

// Usage examples
<Badge>Default</Badge>
<Badge variant="success">Active</Badge>
<Badge variant="warning">Pending</Badge>
<Badge variant="error">Expired</Badge>
```

### Avatar

```tsx
// components/ui/avatar.tsx
export function Avatar({
  src,
  alt,
  fallback,
  size = 'default',
  status,
}: AvatarProps) {
  const sizeClasses = {
    sm: 'h-8 w-8 text-xs',
    default: 'h-10 w-10 text-sm',
    lg: 'h-12 w-12 text-base',
  };

  return (
    <div className="relative">
      {src ? (
        <img
          src={src}
          alt={alt}
          className={cn(
            'rounded-full object-cover',
            sizeClasses[size]
          )}
        />
      ) : (
        <div
          className={cn(
            'rounded-full bg-gray-200 flex items-center justify-center font-medium text-gray-600',
            sizeClasses[size]
          )}
        >
          {fallback}
        </div>
      )}
      {status && (
        <span
          className={cn(
            'absolute bottom-0 right-0 block rounded-full ring-2 ring-white',
            size === 'sm' ? 'h-2 w-2' : 'h-3 w-3',
            status === 'online' && 'bg-green-500',
            status === 'offline' && 'bg-gray-400',
            status === 'busy' && 'bg-red-500'
          )}
        />
      )}
    </div>
  );
}
```

---

## Layout Components

### Page Container

```tsx
// components/layouts/PageContainer.tsx
interface PageContainerProps {
  title: string;
  description?: string;
  action?: React.ReactNode;
  children: React.ReactNode;
}

export function PageContainer({
  title,
  description,
  action,
  children,
}: PageContainerProps) {
  return (
    <div className="flex flex-col min-h-full">
      {/* Page header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">{title}</h1>
          {description && (
            <p className="mt-1 text-sm text-gray-500">{description}</p>
          )}
        </div>
        {action && <div>{action}</div>}
      </div>

      {/* Page content */}
      <div className="flex-1">{children}</div>
    </div>
  );
}

// Usage
<PageContainer
  title="Links"
  description="Manage your shortened links"
  action={<Button>Create Link</Button>}
>
  <LinkList />
</PageContainer>
```

### Sidebar Navigation

```tsx
// components/layouts/Sidebar.tsx
const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Links', href: '/dashboard/links', icon: Link2 },
  { name: 'Analytics', href: '/dashboard/analytics', icon: BarChart3 },
  { name: 'Domains', href: '/dashboard/domains', icon: Globe },
  { name: 'QR Codes', href: '/dashboard/qr-codes', icon: QrCode },
  { name: 'Bio Pages', href: '/dashboard/bio', icon: User },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="w-64 border-r border-gray-200 bg-white">
      {/* Logo */}
      <div className="flex h-16 items-center px-6 border-b border-gray-200">
        <Link href="/dashboard" className="flex items-center gap-2">
          <Logo className="h-8 w-8" />
          <span className="font-semibold text-gray-900">Linkrift</span>
        </Link>
      </div>

      {/* Navigation */}
      <nav className="flex flex-col gap-1 p-4">
        {navigation.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.name}
              href={item.href}
              className={cn(
                'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-brand-50 text-brand-700'
                  : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
              )}
            >
              <item.icon className="h-5 w-5" />
              {item.name}
            </Link>
          );
        })}
      </nav>

      {/* Workspace selector */}
      <div className="absolute bottom-0 left-0 right-0 p-4 border-t border-gray-200">
        <WorkspaceSelector />
      </div>
    </aside>
  );
}
```

### Card

```tsx
// components/ui/card.tsx
export function Card({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        'rounded-lg border border-gray-200 bg-white shadow-sm',
        className
      )}
      {...props}
    />
  );
}

export function CardHeader({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn('flex flex-col space-y-1.5 p-6', className)}
      {...props}
    />
  );
}

export function CardTitle({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h3
      className={cn('text-lg font-semibold text-gray-900', className)}
      {...props}
    />
  );
}

export function CardContent({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('p-6 pt-0', className)} {...props} />;
}

export function CardFooter({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn('flex items-center p-6 pt-0', className)}
      {...props}
    />
  );
}
```

---

## Form Components

### Input

```tsx
// components/ui/input.tsx
export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  error?: string;
  icon?: React.ReactNode;
}

export function Input({ className, error, icon, ...props }: InputProps) {
  return (
    <div className="relative">
      {icon && (
        <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
          {icon}
        </div>
      )}
      <input
        className={cn(
          'flex h-10 w-full rounded-md border bg-white px-3 py-2 text-sm',
          'placeholder:text-gray-400',
          'focus:outline-none focus:ring-2 focus:ring-offset-0',
          'disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-500',
          'transition-colors duration-150',
          icon && 'pl-10',
          error
            ? 'border-red-500 focus:border-red-500 focus:ring-red-500/20'
            : 'border-gray-300 focus:border-brand-500 focus:ring-brand-500/20',
          className
        )}
        {...props}
      />
      {error && (
        <p className="mt-1.5 text-sm text-red-500">{error}</p>
      )}
    </div>
  );
}
```

### Select

```tsx
// components/ui/select.tsx (Radix-based)
export function Select({
  value,
  onValueChange,
  options,
  placeholder = 'Select...',
}: SelectProps) {
  return (
    <SelectPrimitive.Root value={value} onValueChange={onValueChange}>
      <SelectPrimitive.Trigger
        className={cn(
          'flex h-10 w-full items-center justify-between rounded-md border border-gray-300 bg-white px-3 py-2 text-sm',
          'focus:outline-none focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500',
          'disabled:cursor-not-allowed disabled:bg-gray-50'
        )}
      >
        <SelectPrimitive.Value placeholder={placeholder} />
        <SelectPrimitive.Icon>
          <ChevronDown className="h-4 w-4 text-gray-400" />
        </SelectPrimitive.Icon>
      </SelectPrimitive.Trigger>

      <SelectPrimitive.Portal>
        <SelectPrimitive.Content
          className={cn(
            'relative z-50 min-w-[8rem] overflow-hidden rounded-md border border-gray-200 bg-white shadow-lg',
            'animate-in fade-in-0 zoom-in-95'
          )}
        >
          <SelectPrimitive.Viewport className="p-1">
            {options.map((option) => (
              <SelectPrimitive.Item
                key={option.value}
                value={option.value}
                className={cn(
                  'relative flex cursor-pointer select-none items-center rounded-sm py-2 px-3 text-sm',
                  'focus:bg-gray-100 focus:outline-none',
                  'data-[disabled]:pointer-events-none data-[disabled]:opacity-50'
                )}
              >
                <SelectPrimitive.ItemText>
                  {option.label}
                </SelectPrimitive.ItemText>
              </SelectPrimitive.Item>
            ))}
          </SelectPrimitive.Viewport>
        </SelectPrimitive.Content>
      </SelectPrimitive.Portal>
    </SelectPrimitive.Root>
  );
}
```

### Form Field

```tsx
// components/forms/FormField.tsx
interface FormFieldProps {
  label: string;
  description?: string;
  error?: string;
  required?: boolean;
  children: React.ReactNode;
}

export function FormField({
  label,
  description,
  error,
  required,
  children,
}: FormFieldProps) {
  return (
    <div className="space-y-2">
      <label className="block">
        <span className="text-sm font-medium text-gray-700">
          {label}
          {required && <span className="text-red-500 ml-0.5">*</span>}
        </span>
        {description && (
          <span className="block text-sm text-gray-500 mt-0.5">
            {description}
          </span>
        )}
      </label>
      {children}
      {error && (
        <p className="text-sm text-red-500" role="alert">
          {error}
        </p>
      )}
    </div>
  );
}
```

---

## Data Display

### Data Table

```tsx
// components/ui/data-table.tsx
interface Column<T> {
  key: keyof T | string;
  header: string;
  render?: (row: T) => React.ReactNode;
  sortable?: boolean;
  width?: string;
}

interface DataTableProps<T> {
  data: T[];
  columns: Column<T>[];
  onRowClick?: (row: T) => void;
  loading?: boolean;
  emptyState?: React.ReactNode;
}

export function DataTable<T extends { id: string }>({
  data,
  columns,
  onRowClick,
  loading,
  emptyState,
}: DataTableProps<T>) {
  if (loading) {
    return <TableSkeleton columns={columns.length} rows={5} />;
  }

  if (data.length === 0 && emptyState) {
    return <>{emptyState}</>;
  }

  return (
    <div className="overflow-hidden rounded-lg border border-gray-200">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            {columns.map((column) => (
              <th
                key={column.key as string}
                className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                style={{ width: column.width }}
              >
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-100">
          {data.map((row) => (
            <tr
              key={row.id}
              onClick={() => onRowClick?.(row)}
              className={cn(
                'transition-colors',
                onRowClick && 'cursor-pointer hover:bg-gray-50'
              )}
            >
              {columns.map((column) => (
                <td
                  key={column.key as string}
                  className="px-4 py-3 text-sm text-gray-700"
                >
                  {column.render
                    ? column.render(row)
                    : (row[column.key as keyof T] as React.ReactNode)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

### Stats Card

```tsx
// components/features/analytics/StatsCard.tsx
interface StatsCardProps {
  title: string;
  value: string | number;
  change?: {
    value: number;
    type: 'increase' | 'decrease';
  };
  icon?: React.ReactNode;
  loading?: boolean;
}

export function StatsCard({
  title,
  value,
  change,
  icon,
  loading,
}: StatsCardProps) {
  if (loading) {
    return (
      <Card className="p-6">
        <div className="animate-pulse space-y-3">
          <div className="h-4 bg-gray-200 rounded w-1/3" />
          <div className="h-8 bg-gray-200 rounded w-1/2" />
        </div>
      </Card>
    );
  }

  return (
    <Card className="p-6">
      <div className="flex items-center justify-between">
        <p className="text-sm font-medium text-gray-500">{title}</p>
        {icon && <div className="text-gray-400">{icon}</div>}
      </div>
      <div className="mt-2 flex items-baseline gap-2">
        <p className="text-3xl font-semibold text-gray-900">
          {typeof value === 'number' ? value.toLocaleString() : value}
        </p>
        {change && (
          <span
            className={cn(
              'inline-flex items-center text-sm font-medium',
              change.type === 'increase' ? 'text-green-600' : 'text-red-600'
            )}
          >
            {change.type === 'increase' ? (
              <TrendingUp className="h-4 w-4 mr-0.5" />
            ) : (
              <TrendingDown className="h-4 w-4 mr-0.5" />
            )}
            {Math.abs(change.value)}%
          </span>
        )}
      </div>
    </Card>
  );
}
```

---

## Feedback Components

### Toast

```tsx
// components/ui/toast.tsx
const toastVariants = cva(
  'flex items-center gap-3 px-4 py-3 rounded-lg shadow-lg border',
  {
    variants: {
      variant: {
        default: 'bg-white border-gray-200 text-gray-900',
        success: 'bg-green-50 border-green-200 text-green-800',
        error: 'bg-red-50 border-red-200 text-red-800',
        warning: 'bg-amber-50 border-amber-200 text-amber-800',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
);

export function Toast({ variant, title, description, action }: ToastProps) {
  return (
    <div className={cn(toastVariants({ variant }))}>
      <div className="flex-1">
        {title && <p className="font-medium">{title}</p>}
        {description && (
          <p className="text-sm opacity-90">{description}</p>
        )}
      </div>
      {action}
    </div>
  );
}
```

### Dialog

```tsx
// components/ui/dialog.tsx
export function Dialog({ open, onOpenChange, children }: DialogProps) {
  return (
    <DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay
          className={cn(
            'fixed inset-0 z-50 bg-black/50',
            'data-[state=open]:animate-in data-[state=closed]:animate-out',
            'data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0'
          )}
        />
        <DialogPrimitive.Content
          className={cn(
            'fixed left-[50%] top-[50%] z-50 w-full max-w-lg translate-x-[-50%] translate-y-[-50%]',
            'rounded-lg border border-gray-200 bg-white p-6 shadow-xl',
            'data-[state=open]:animate-in data-[state=closed]:animate-out',
            'data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
            'data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95',
            'data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%]',
            'data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%]'
          )}
        >
          {children}
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}

export function DialogHeader({ children }: { children: React.ReactNode }) {
  return <div className="mb-4">{children}</div>;
}

export function DialogTitle({ children }: { children: React.ReactNode }) {
  return (
    <DialogPrimitive.Title className="text-lg font-semibold text-gray-900">
      {children}
    </DialogPrimitive.Title>
  );
}

export function DialogFooter({ children }: { children: React.ReactNode }) {
  return (
    <div className="mt-6 flex justify-end gap-3">{children}</div>
  );
}
```

---

## Dashboard Components

### Link Card

```tsx
// components/features/links/LinkCard.tsx
export function LinkCard({ link, onEdit, onDelete }: LinkCardProps) {
  const { toast } = useToast();

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(link.short_url);
    toast({ title: 'Copied to clipboard' });
  };

  return (
    <Card className="group hover:shadow-md transition-shadow">
      <CardContent className="p-4">
        <div className="flex items-start gap-4">
          {/* Favicon */}
          <div className="flex-shrink-0">
            {link.favicon_url ? (
              <img
                src={link.favicon_url}
                alt=""
                className="h-10 w-10 rounded-lg object-cover"
              />
            ) : (
              <div className="h-10 w-10 rounded-lg bg-gray-100 flex items-center justify-center">
                <Link2 className="h-5 w-5 text-gray-400" />
              </div>
            )}
          </div>

          {/* Content */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <a
                href={link.short_url}
                target="_blank"
                rel="noopener noreferrer"
                className="font-mono text-brand-600 hover:text-brand-700 font-medium"
              >
                {link.short_code}
              </a>
              <button
                onClick={copyToClipboard}
                className="opacity-0 group-hover:opacity-100 transition-opacity p-1 hover:bg-gray-100 rounded"
              >
                <Copy className="h-4 w-4 text-gray-400" />
              </button>
            </div>

            <p className="text-sm text-gray-600 truncate mt-1">
              {link.title || link.url}
            </p>

            <div className="flex items-center gap-4 mt-2 text-xs text-gray-500">
              <span className="flex items-center gap-1">
                <BarChart2 className="h-3.5 w-3.5" />
                {link.total_clicks.toLocaleString()} clicks
              </span>
              <span>
                {formatDistanceToNow(new Date(link.created_at), {
                  addSuffix: true,
                })}
              </span>
            </div>
          </div>

          {/* Actions */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="opacity-0 group-hover:opacity-100">
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onEdit?.(link)}>
                <Pencil className="h-4 w-4 mr-2" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem onClick={copyToClipboard}>
                <Copy className="h-4 w-4 mr-2" />
                Copy URL
              </DropdownMenuItem>
              <DropdownMenuItem>
                <QrCode className="h-4 w-4 mr-2" />
                QR Code
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="text-red-600"
                onClick={() => onDelete?.(link)}
              >
                <Trash2 className="h-4 w-4 mr-2" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardContent>
    </Card>
  );
}
```

---

## Loading States

### Skeleton

```tsx
// components/ui/skeleton.tsx
export function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        'animate-pulse rounded-md bg-gray-200',
        className
      )}
      {...props}
    />
  );
}

// Link Card Skeleton
export function LinkCardSkeleton() {
  return (
    <Card className="p-4">
      <div className="flex items-start gap-4">
        <Skeleton className="h-10 w-10 rounded-lg" />
        <div className="flex-1 space-y-2">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-4 w-48" />
          <Skeleton className="h-3 w-32" />
        </div>
      </div>
    </Card>
  );
}

// Table Skeleton
export function TableSkeleton({ columns, rows }: { columns: number; rows: number }) {
  return (
    <div className="overflow-hidden rounded-lg border border-gray-200">
      <div className="bg-gray-50 px-4 py-3">
        <div className="flex gap-4">
          {Array.from({ length: columns }).map((_, i) => (
            <Skeleton key={i} className="h-4 w-24" />
          ))}
        </div>
      </div>
      <div className="divide-y divide-gray-100">
        {Array.from({ length: rows }).map((_, i) => (
          <div key={i} className="px-4 py-3">
            <div className="flex gap-4">
              {Array.from({ length: columns }).map((_, j) => (
                <Skeleton
                  key={j}
                  className="h-4"
                  style={{ width: `${Math.random() * 40 + 60}px` }}
                />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
```

---

## Empty States

```tsx
// components/ui/empty-state.tsx
interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
}

export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
      {icon && (
        <div className="mb-4 rounded-full bg-gray-100 p-3">
          <div className="h-6 w-6 text-gray-400">{icon}</div>
        </div>
      )}
      <h3 className="text-sm font-medium text-gray-900">{title}</h3>
      {description && (
        <p className="mt-1 text-sm text-gray-500 max-w-sm">{description}</p>
      )}
      {action && (
        <Button onClick={action.onClick} className="mt-4">
          {action.label}
        </Button>
      )}
    </div>
  );
}

// Usage examples
<EmptyState
  icon={<Link2 />}
  title="No links yet"
  description="Create your first short link to get started"
  action={{
    label: 'Create Link',
    onClick: () => setCreateDialogOpen(true),
  }}
/>

<EmptyState
  icon={<Search />}
  title="No results found"
  description="Try adjusting your search or filters"
/>
```

---

## Related Documentation

- [Design System](DESIGN_SYSTEM.md) — Design principles
- [Frontend Architecture](../architecture/FRONTEND_ARCHITECTURE.md) — React patterns
- [Accessibility](../reference/ACCESSIBILITY.md) — WCAG compliance
