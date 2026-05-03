/**
 * Label–value row used inside the candidate profile section.
 * Caption-strong uppercase label on the left (28-rem column), value
 * on the right with `min-w-0 break-words` so long PDF-extracted strings
 * don't blow out the card.
 */
export function ProfileLine({
  label,
  value,
}: {
  label: string
  value: string
}) {
  return (
    <div className="flex items-baseline gap-3">
      <span className="text-caption text-muted w-28 shrink-0 uppercase">
        {label}
      </span>
      <span className="text-body-md text-ink min-w-0 break-words">{value}</span>
    </div>
  )
}
