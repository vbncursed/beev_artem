import { useI18n } from '@/app/providers/I18nProvider'
import { Card } from '@/presentation/ui'
import { H2, H3, LegalLayout, P } from './LegalLayout'

export function SupportPage() {
  const { locale, t } = useI18n()
  return (
    <LegalLayout
      eyebrow={t('legal.eyebrow.support')}
      title={t('legal.support.title')}
      subtitle={t('legal.support.subtitle')}
    >
      {locale === 'ru' ? <BodyRu /> : <BodyEn />}
    </LegalLayout>
  )
}

function BodyRu() {
  return (
    <>
      <ContactCard
        email="support@cadence.example"
        hoursLabel="Часы работы"
        hoursValue="Пн–Пт, 09:00–19:00 МСК"
        responseLabel="Время ответа"
        responseValue="≤ 1 рабочего дня"
      />

      <H2>Частые вопросы</H2>

      <H3>Как загрузить резюме кандидата?</H3>
      <P>
        Откройте вакансию → drag&drop файла в зону «Перетащите резюме».
        Принимаются PDF, DOC, DOCX, TXT до 10 МБ. Cadence автоматически
        создаст карточку кандидата, извлечёт профиль и запустит анализ.
      </P>

      <H3>Что означает match score?</H3>
      <P>
        Score (0–100) — результат эвристики, которая считает совпадение
        навыков кандидата с требованиями вакансии. Веса навыков и флаги
        must-have / nice-to-have задаются при создании вакансии. Подробный
        разбор виден на странице кандидата в секции «Разбор навыков».
      </P>

      <H3>Почему AI-вердикт может отличаться от score?</H3>
      <P>
        Score детерминирован и основан на ключевых словах. AI (Yandex Cloud
        Foundation Models) видит контекст резюме и может скорректировать
        вердикт — например, поднять «Hire» для кандидата с релевантным
        опытом, чьи навыки названы по-другому. Решение всегда за HR.
      </P>

      <H3>Можно ли скачать оригинальный файл резюме?</H3>
      <P>
        Да. Откройте кандидата → справа в панели анализа есть кнопка
        «Скачать резюме». Файл вернётся в исходном формате.
      </P>

      <H3>Как удалить кандидата?</H3>
      <P>
        В правой панели анализа кнопка «Удалить кандидата» с подтверждением.
        Удаление каскадно убирает резюме и метаданные. Аналитические записи
        остаются в исторических целях, но из списков кандидата исчезают.
      </P>

      <H3>Что происходит, если LLM недоступна?</H3>
      <P>
        Эвристический score вычисляется всегда — он не зависит от LLM. Если
        Yandex API временно недоступен, кандидат всё равно получит score и
        статус «Готово»; AI-обоснование будет содержать дефолтный текст
        («Эвристическая оценка»). Можно переанализировать кандидата
        позже, когда LLM восстановится.
      </P>

      <H2>Документация</H2>
      <P>
        Спецификация API живёт по адресу{' '}
        <code className="rounded-xs bg-surface-strong px-1.5 py-0.5 font-mono text-sm">
          {window.location.origin}/swagger.json
        </code>
        ; интерактивный UI —{' '}
        <a
          href="/docs"
          className="cursor-pointer text-primary hover:opacity-80"
        >
          /docs
        </a>
        . Технические README сервисов опубликованы в репозитории.
      </P>
    </>
  )
}

function BodyEn() {
  return (
    <>
      <ContactCard
        email="support@cadence.example"
        hoursLabel="Hours"
        hoursValue="Mon–Fri, 09:00–19:00 MSK"
        responseLabel="Response time"
        responseValue="≤ 1 business day"
      />

      <H2>Frequently asked questions</H2>

      <H3>How do I upload a candidate's resume?</H3>
      <P>
        Open a vacancy → drag &amp; drop the file into the "Drop a resume"
        zone. PDF, DOC, DOCX, TXT up to 10 MB are accepted. Cadence creates
        the candidate, extracts the profile, and runs analysis automatically.
      </P>

      <H3>What does the match score mean?</H3>
      <P>
        The score (0–100) is the output of the heuristic Scorer, matching
        candidate skills against vacancy requirements. Weights and
        must-have / nice-to-have flags are set when creating the vacancy.
        Full breakdown is on the candidate panel under "Skills breakdown".
      </P>

      <H3>Why might the AI verdict differ from the score?</H3>
      <P>
        The score is deterministic and keyword-based. The AI (Yandex Cloud
        Foundation Models) reads the full resume context and may adjust the
        verdict — for example, recommending "Hire" for a candidate whose
        relevant skills are phrased differently. The final call is always
        the recruiter's.
      </P>

      <H3>Can I download the original resume file?</H3>
      <P>
        Yes. Open the candidate → "Download resume" in the analysis panel.
        The file is returned in its original format.
      </P>

      <H3>How do I remove a candidate?</H3>
      <P>
        In the analysis panel, click "Remove candidate" and confirm.
        Deletion cascades to the resume and metadata. Analytical records are
        kept for history but disappear from candidate lists.
      </P>

      <H3>What happens if the LLM is unavailable?</H3>
      <P>
        The heuristic score is always computed — it does not depend on the
        LLM. If the Yandex API is temporarily down, the candidate still gets
        a score and "Done" status; the AI rationale will contain default
        text. You can re-analyze later when the LLM recovers.
      </P>

      <H2>Documentation</H2>
      <P>
        The API spec lives at{' '}
        <code className="rounded-xs bg-surface-strong px-1.5 py-0.5 font-mono text-sm">
          {window.location.origin}/swagger.json
        </code>
        ; interactive UI at{' '}
        <a
          href="/docs"
          className="cursor-pointer text-primary hover:opacity-80"
        >
          /docs
        </a>
        . Service READMEs are in the repository.
      </P>
    </>
  )
}

function ContactCard({
  email,
  hoursLabel,
  hoursValue,
  responseLabel,
  responseValue,
}: {
  email: string
  hoursLabel: string
  hoursValue: string
  responseLabel: string
  responseValue: string
}) {
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
