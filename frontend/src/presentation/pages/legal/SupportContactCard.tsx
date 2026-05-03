import { Card } from '@/presentation/ui'

type Props = {
  email: string
  hoursLabel: string
  hoursValue: string
  responseLabel: string
  responseValue: string
}

export function SupportContactCard({
  email,
  hoursLabel,
  hoursValue,
  responseLabel,
  responseValue,
}: Props) {
  return (
    <Card variant="feature" elevated className="grid gap-6 md:grid-cols-2">
      <div>
        <p className="text-caption-strong text-muted uppercase">Email</p>
        <a
          href={`mailto:${email}`}
          className="text-title-md text-primary mt-2 inline-block cursor-pointer hover:opacity-80"
        >
          {email}
        </a>
      </div>
      <div className="flex flex-col gap-3">
        <Row label={hoursLabel} value={hoursValue} />
        <Row label={responseLabel} value={responseValue} />
      </div>
    </Card>
  )
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-caption-strong text-muted uppercase">{label}</p>
      <p className="text-body-md text-ink mt-1">{value}</p>
    </div>
  )
}
