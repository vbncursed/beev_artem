import { useI18n } from '@/app/providers/I18nProvider'
import { H2, LegalLayout, P, UL } from './LegalLayout'

export function TermsPage() {
  const { locale, t } = useI18n()
  return (
    <LegalLayout
      eyebrow={t('legal.eyebrow.terms')}
      title={t('legal.terms.title')}
      subtitle={t('legal.terms.subtitle')}
    >
      {locale === 'ru' ? <BodyRu /> : <BodyEn />}
    </LegalLayout>
  )
}

function BodyRu() {
  return (
    <>
      <H2>1. Принятие условий</H2>
      <P>
        Создавая аккаунт на Cadence, вы соглашаетесь с этими условиями. Если
        вы не согласны хотя бы с одним пунктом — не используйте сервис.
        Условия могут меняться; существенные изменения публикуются за 30
        дней до вступления в силу.
      </P>

      <H2>2. Аккаунт и ответственность</H2>
      <UL>
        <li>один аккаунт — одно физическое лицо; передача аккаунта запрещена;</li>
        <li>пароль должен быть длиной не менее 8 символов; рекомендуем парольный менеджер;</li>
        <li>вы несёте ответственность за все действия, совершённые из вашего аккаунта;</li>
        <li>о подозрительной активности сообщайте на support@cadence.example.</li>
      </UL>

      <H2>3. Допустимое использование</H2>
      <P>
        Сервис предназначен для подбора персонала на собственные вакансии вашей
        компании. Разрешено:
      </P>
      <UL>
        <li>создавать вакансии и описывать требуемые навыки;</li>
        <li>загружать резюме кандидатов, которые сами согласились на обработку;</li>
        <li>анализировать соответствие через эвристику и AI;</li>
        <li>экспортировать собственные данные в любой момент.</li>
      </UL>

      <H2>4. Запрещено</H2>
      <UL>
        <li>загружать резюме без согласия кандидата;</li>
        <li>перепродавать данные кандидатов третьим лицам;</li>
        <li>использовать сервис как базу данных кандидатов для перепродажи;</li>
        <li>автоматизированный массовый scraping или DDoS;</li>
        <li>обход системы аутентификации, попытки эксплуатации уязвимостей;</li>
        <li>загрузка вредоносного контента (вирусы, фишинг).</li>
      </UL>

      <H2>5. Содержимое и интеллектуальная собственность</H2>
      <P>
        Вы сохраняете все права на загружаемые резюме и созданные вакансии.
        Cadence получает ограниченную лицензию на хранение и обработку этих
        данных исключительно для оказания услуги. Логотип, дизайн и код
        Cadence принадлежат нам.
      </P>

      <H2>6. AI-вердикт — рекомендация, а не решение</H2>
      <P>
        Score и AI-обоснование являются справочной информацией. Решение о
        найме принимаете вы. Cadence не несёт ответственности за решения,
        принятые на основе AI-рекомендаций. AI может ошибаться — особенно на
        нестандартных резюме.
      </P>

      <H2>7. Доступность сервиса</H2>
      <P>
        Мы стремимся к высокой доступности, но не гарантируем 100% uptime.
        Плановые работы анонсируем заранее. SLA для разных тарифов описаны в
        отдельном договоре.
      </P>

      <H2>8. Прекращение</H2>
      <P>
        Вы можете удалить аккаунт в любой момент. Мы можем заблокировать
        доступ при нарушении этих условий — обычно с предварительным
        уведомлением, кроме случаев явных угроз безопасности других
        пользователей.
      </P>

      <H2>9. Ограничение ответственности</H2>
      <P>
        Cadence предоставляется «как есть». В пределах, разрешённых законом,
        мы не несём ответственности за упущенную выгоду, репутационные потери
        или косвенный ущерб. Максимальная совокупная ответственность
        ограничена суммой, оплаченной вами за последние 12 месяцев.
      </P>

      <H2>10. Применимое право</H2>
      <P>
        Условия регулируются законодательством Российской Федерации. Споры
        рассматриваются по месту регистрации Cadence LLC.
      </P>
    </>
  )
}

function BodyEn() {
  return (
    <>
      <H2>1. Acceptance</H2>
      <P>
        By creating an account on Cadence you agree to these terms. If you
        disagree with any clause, do not use the service. Terms may change;
        material changes are posted at least 30 days before they take effect.
      </P>

      <H2>2. Account and responsibility</H2>
      <UL>
        <li>one account per natural person; account transfer is prohibited;</li>
        <li>passwords must be at least 8 characters; we recommend a password manager;</li>
        <li>you are responsible for all activity under your account;</li>
        <li>report suspicious activity to support@cadence.example.</li>
      </UL>

      <H2>3. Permitted use</H2>
      <P>
        The service is intended for hiring against your own company's
        vacancies. You may:
      </P>
      <UL>
        <li>create vacancies and describe required skills;</li>
        <li>upload résumés of candidates who consented to processing;</li>
        <li>analyse fit via heuristic and AI;</li>
        <li>export your own data at any time.</li>
      </UL>

      <H2>4. Prohibited</H2>
      <UL>
        <li>uploading résumés without the candidate's consent;</li>
        <li>reselling candidate data to third parties;</li>
        <li>using the service as a candidate database for resale;</li>
        <li>automated mass scraping or DDoS;</li>
        <li>bypassing authentication or attempting to exploit vulnerabilities;</li>
        <li>uploading malicious content (viruses, phishing).</li>
      </UL>

      <H2>5. Content and IP</H2>
      <P>
        You retain all rights to résumés and vacancies you upload. Cadence
        receives a limited licence to store and process this data solely to
        deliver the service. The Cadence logo, design and code belong to us.
      </P>

      <H2>6. AI verdict — recommendation, not decision</H2>
      <P>
        The match score and AI rationale are advisory. Hiring decisions are
        yours. Cadence is not liable for decisions made on the basis of AI
        recommendations. AI can be wrong — especially on non-standard
        résumés.
      </P>

      <H2>7. Availability</H2>
      <P>
        We strive for high availability but do not guarantee 100% uptime.
        Planned maintenance is announced in advance. SLAs for different
        plans are covered in a separate contract.
      </P>

      <H2>8. Termination</H2>
      <P>
        You may delete your account at any time. We may suspend access for
        violation of these terms — usually with prior notice, except in cases
        of clear threat to other users' security.
      </P>

      <H2>9. Limitation of liability</H2>
      <P>
        Cadence is provided "as is". To the extent permitted by law, we are
        not liable for lost profits, reputational damage or indirect harm.
        Maximum aggregate liability is capped at the amount you paid in the
        prior 12 months.
      </P>

      <H2>10. Governing law</H2>
      <P>
        These terms are governed by the laws of the Russian Federation.
        Disputes are resolved at the seat of Cadence LLC.
      </P>
    </>
  )
}
