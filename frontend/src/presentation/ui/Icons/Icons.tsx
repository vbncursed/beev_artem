import { cn } from '@/shared/lib/cn'

/**
 * Shared icon set. Only icons used in 2+ places (or genuinely reusable
 * primitives like `IconBox` chrome) live here — per the principle of not
 * extracting before the duplication pain. Component-bound icons (search
 * input glass, theme toggle sun/moon, language flags, brand mark) stay
 * inline next to their owner.
 *
 * All icons share the same SVG defaults: stroke-only outline, 2px stroke,
 * `currentColor` so `text-*` utilities pick up the colour. Default size
 * 16px; pass a width/height override via `size` prop where needed.
 */

type IconProps = {
  size?: number
  className?: string
  title?: string
}

const SVG_DEFAULTS = {
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 2,
  strokeLinecap: 'round' as const,
  strokeLinejoin: 'round' as const,
}

function svgProps({ size = 16, className, title }: IconProps) {
  return {
    width: size,
    height: size,
    viewBox: '0 0 24 24',
    className: cn('shrink-0', className),
    'aria-hidden': title === undefined,
    role: title !== undefined ? ('img' as const) : undefined,
    ...SVG_DEFAULTS,
  }
}

export function TrashIcon(props: IconProps) {
  return (
    <svg {...svgProps(props)}>
      {props.title && <title>{props.title}</title>}
      <path d="M3 6h18" />
      <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
      <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6" />
      <path d="M10 11v6" />
      <path d="M14 11v6" />
    </svg>
  )
}

export function DownloadIcon(props: IconProps) {
  return (
    <svg {...svgProps(props)}>
      {props.title && <title>{props.title}</title>}
      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
      <path d="M7 10l5 5 5-5" />
      <path d="M12 15V3" />
    </svg>
  )
}

export function ChevronDownIcon(props: IconProps) {
  return (
    <svg {...svgProps({ ...props, size: props.size ?? 12 })}>
      {props.title && <title>{props.title}</title>}
      <path d="m6 9 6 6 6-6" />
    </svg>
  )
}

export function CheckIcon(props: IconProps) {
  return (
    <svg {...svgProps({ ...props, size: props.size ?? 14 })}>
      {props.title && <title>{props.title}</title>}
      <path d="M20 6 9 17l-5-5" />
    </svg>
  )
}
