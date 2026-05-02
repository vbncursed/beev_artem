import { cn } from '@/shared/lib/cn'

export function Spinner({
  size = 16,
  className,
}: {
  size?: number
  className?: string
}) {
  return (
    <span
      role="status"
      aria-label="Loading"
      className={cn(
        'inline-block animate-spin rounded-full border-2 border-current border-t-transparent text-primary',
        className,
      )}
      style={{ width: size, height: size }}
    />
  )
}
