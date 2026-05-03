import { useI18n } from '@/app/providers/I18nProvider'
import { H2, H3, LegalLayout, P, UL } from './LegalLayout'

export function PrivacyPage() {
  const { locale, t } = useI18n()
  return (
    <LegalLayout
      eyebrow={t('legal.eyebrow.privacy')}
      title={t('legal.privacy.title')}
      subtitle={t('legal.privacy.subtitle')}
    >
      {locale === 'ru' ? <BodyRu /> : <BodyEn />}
    </LegalLayout>
  )
}

function BodyRu() {
  return (
    <>
      <H2>1. Кто мы и за что отвечаем</H2>
      <P>
        Cadence — институциональная платформа для подбора персонала. Сервис
        предоставляется ООО «Cadence» (далее — «мы», «нам», «нас»). Эта
        политика описывает, какие персональные данные мы обрабатываем, как
        и с какой целью.
      </P>

      <H2>2. Какие данные мы собираем</H2>
      <H3>От HR-пользователей</H3>
      <UL>
        <li>email-адрес и пароль (хранится в виде bcrypt-хеша);</li>
        <li>метаданные сессий (IP, User-Agent, временные метки);</li>
        <li>содержимое создаваемых вакансий и навыков.</li>
      </UL>
      <H3>От кандидатов (загружаются HR-пользователями)</H3>
      <UL>
        <li>файл резюме (PDF / DOCX / TXT) и его извлечённый текст;</li>
        <li>имя, email, номер телефона — извлекаются из резюме автоматически;</li>
        <li>результаты автоматической оценки (score, обоснование AI).</li>
      </UL>

      <H2>3. Цели обработки</H2>
      <UL>
        <li>аутентификация и защита аккаунтов HR-пользователей;</li>
        <li>сопоставление кандидатов с открытыми вакансиями;</li>
        <li>генерация AI-обоснованной рекомендации по найму;</li>
        <li>аудит решений и техническое обеспечение работы сервиса.</li>
      </UL>

      <H2>4. Передача третьим лицам</H2>
      <P>
        Текст резюме и описание вакансии передаются в Yandex Cloud Foundation
        Models для генерации AI-вердикта. Yandex обрабатывает эти данные
        согласно своей{' '}
        <a
          href="https://yandex.ru/legal/cloud_termsofuse/"
          target="_blank"
          rel="noopener noreferrer"
          className="cursor-pointer text-primary hover:opacity-80"
        >
          политике обработки
        </a>
        . Никаким другим третьим сторонам данные не передаются.
      </P>

      <H2>5. Где и как мы храним данные</H2>
      <P>
        Все данные хранятся в PostgreSQL на серверах в Российской Федерации.
        Файлы резюме хранятся в зашифрованной БД как BYTEA. Резервные копии
        делаются ежедневно и хранятся 30 дней.
      </P>

      <H2>6. Сроки хранения</H2>
      <UL>
        <li>аккаунты HR — до удаления самим пользователем или прекращения договора;</li>
        <li>кандидаты — до удаления HR-пользователем (кнопка «Удалить кандидата»);</li>
        <li>сессионные токены — до истечения TTL (access 15 минут, refresh 30 дней) или явного logout.</li>
      </UL>

      <H2>7. Ваши права</H2>
      <P>
        Вы имеете право: запросить копию своих данных, потребовать их удаления
        или исправления, отозвать согласие на обработку. Запросы отправляйте
        на{' '}
        <a
          href="mailto:privacy@cadence.example"
          className="cursor-pointer text-primary hover:opacity-80"
        >
          privacy@cadence.example
        </a>
        . Срок ответа — до 30 дней.
      </P>

      <H2>8. Безопасность</H2>
      <P>
        Доступ к данным аутентифицирован через JWT, каждый сервис повторно
        валидирует токен (defense in depth). Пароли хранятся через bcrypt
        cost 12. Внутренние gRPC-каналы изолированы в private docker-сети.
      </P>

      <H2>9. Изменения политики</H2>
      <P>
        Существенные изменения публикуются на этой странице за 30 дней до
        вступления в силу. Дата последнего обновления указана в шапке.
      </P>
    </>
  )
}

function BodyEn() {
  return (
    <>
      <H2>1. Who we are</H2>
      <P>
        Cadence is an institutional hiring platform. The service is operated
        by Cadence LLC (hereafter "we", "us"). This policy describes what
        personal data we process, how and why.
      </P>

      <H2>2. What we collect</H2>
      <H3>From HR users</H3>
      <UL>
        <li>email address and password (stored as a bcrypt hash);</li>
        <li>session metadata (IP, User-Agent, timestamps);</li>
        <li>vacancy content and skill weights you create.</li>
      </UL>
      <H3>From candidates (uploaded by HR users)</H3>
      <UL>
        <li>resume file (PDF / DOCX / TXT) and its extracted text;</li>
        <li>name, email, phone — auto-extracted from the resume;</li>
        <li>scoring output (match score, AI rationale).</li>
      </UL>

      <H2>3. Purposes of processing</H2>
      <UL>
        <li>authentication and account protection for HR users;</li>
        <li>matching candidates against open vacancies;</li>
        <li>generating AI-backed hiring recommendations;</li>
        <li>auditing decisions and operating the service.</li>
      </UL>

      <H2>4. Third-party processors</H2>
      <P>
        Resume text and vacancy descriptions are sent to Yandex Cloud
        Foundation Models for AI verdict generation. Yandex processes this
        data per its{' '}
        <a
          href="https://yandex.com/legal/cloud_termsofuse/"
          target="_blank"
          rel="noopener noreferrer"
          className="cursor-pointer text-primary hover:opacity-80"
        >
          terms of service
        </a>
        . No other third parties receive data.
      </P>

      <H2>5. Storage</H2>
      <P>
        All data is stored in PostgreSQL on servers in the Russian
        Federation. Resume files live in the database as encrypted BYTEA.
        Backups run daily and are retained for 30 days.
      </P>

      <H2>6. Retention</H2>
      <UL>
        <li>HR accounts — until deletion by the user or contract termination;</li>
        <li>candidates — until deletion by the HR user via "Remove candidate";</li>
        <li>session tokens — until TTL expiry (access 15 min, refresh 30 days) or explicit logout.</li>
      </UL>

      <H2>7. Your rights</H2>
      <P>
        You may request a copy of your data, demand its deletion or
        correction, or withdraw consent. Send requests to{' '}
        <a
          href="mailto:privacy@cadence.example"
          className="cursor-pointer text-primary hover:opacity-80"
        >
          privacy@cadence.example
        </a>
        . We respond within 30 days.
      </P>

      <H2>8. Security</H2>
      <P>
        Access is authenticated via JWT, with every service revalidating the
        token (defense in depth). Passwords are hashed with bcrypt cost 12.
        Internal gRPC channels are isolated on a private docker network.
      </P>

      <H2>9. Changes</H2>
      <P>
        Material changes are posted here at least 30 days before they take
        effect. The "last updated" date is shown at the top of the page.
      </P>
    </>
  )
}
