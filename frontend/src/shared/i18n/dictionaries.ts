export type Locale = 'ru' | 'en'

/**
 * Flat key dictionary. Keys group by feature. Variable interpolation
 * supports `{name}` placeholders. Russian is the source of truth for
 * pluralisation (manual `plural()` helper handles 1/few/many forms).
 */
type Dict = Record<string, string>

const ru: Dict = {
  // Common
  'common.signIn': 'Войти',
  'common.signOut': 'Выйти',
  'common.cancel': 'Отмена',
  'common.error': 'Ошибка',
  'common.loading': 'Загрузка',
  'common.back': '← Назад',
  'common.create': 'Создать',
  'common.save': 'Сохранить',
  'common.search': 'Поиск',
  'common.email': 'Email',
  'common.password': 'Пароль',

  // Nav
  'nav.vacancies': 'Вакансии',

  // Theme toggle (a11y)
  'theme.toLight': 'Светлая тема',
  'theme.toDark': 'Тёмная тема',

  // Locale toggle (a11y)
  'locale.label': 'Язык интерфейса',
  'locale.ru': 'Русский',
  'locale.en': 'Английский',

  // Auth page — marketing
  'auth.eyebrow': 'Институциональный найм',
  'auth.titleLogin': 'С возвращением в Cadence.',
  'auth.titleRegister': 'Найм в ровном ритме.',
  'auth.subtitle':
    'Приём вакансий, AI-скоринг резюме и решения, которые можно защитить — на одном спокойном холсте.',
  'auth.modeLogin': 'Войти',
  'auth.modeRegister': 'Регистрация',
  'auth.formTitleLogin': 'Вход',
  'auth.formTitleRegister': 'Создание аккаунта',
  'auth.formHintLogin': 'Используйте рабочий email и пароль.',
  'auth.formHintRegister':
    'Минимум восемь символов. Используйте парольную фразу, которую сможете запомнить.',
  'auth.confirmPassword': 'Подтвердите пароль',
  'auth.passwordHintRegister': 'Не короче 8 символов',
  'auth.submitLogin': 'Войти',
  'auth.submitRegister': 'Создать аккаунт',

  // Auth — validation + server errors
  'auth.error.emailRequired': 'Email обязателен',
  'auth.error.emailInvalid': 'Неверный формат email',
  'auth.error.passwordShort': 'Не короче 8 символов',
  'auth.error.passwordLong': 'Не длиннее 72 символов',
  'auth.error.passwordsNoMatch': 'Пароли не совпадают',
  'auth.error.invalidCreds': 'Неверный email или пароль',
  'auth.error.exists': 'Аккаунт с таким email уже существует',
  'auth.error.invalidArg': 'Проверьте введённые данные',
  'auth.error.rateLimited': 'Слишком много попыток — попробуйте позже',
  'auth.error.network': 'Не удаётся связаться с сервером',
  'auth.error.fallbackLogin': 'Не удалось войти. Попробуйте ещё раз.',
  'auth.error.fallbackRegister':
    'Не удалось зарегистрироваться. Попробуйте ещё раз.',

  // Vacancies list
  'vacancies.eyebrow': 'Воронка',
  'vacancies.title': 'Вакансии',
  'vacancies.subtitle':
    'Открытые позиции, оценённые кандидаты, решения, которые можно защитить.',
  'vacancies.new': 'Новая вакансия',
  'vacancies.searchPlaceholder': 'Поиск по названию, описанию',
  'vacancies.countOne': '{n} вакансия',
  'vacancies.countFew': '{n} вакансии',
  'vacancies.countMany': '{n} вакансий',
  'vacancies.empty.filteredTitle': 'По фильтрам ничего не найдено.',
  'vacancies.empty.filteredHint':
    'Попробуйте другую роль или очистите поиск.',
  'vacancies.empty.title': 'Пока ни одной вакансии.',
  'vacancies.empty.hint': 'Откройте первую позицию и собирайте кандидатов.',
  'vacancies.empty.cta': 'Создать вакансию',
  'vacancies.empty.badgeFiltered': 'Не найдено',
  'vacancies.empty.badge': 'Пусто',
  'vacancies.error.loadFailed': 'Не удалось загрузить вакансии.',
  'vacancies.error.title': 'Не получилось загрузить вакансии.',

  // Roles
  'roles.all': 'Все роли',
  'roles.accountant': 'Бухгалтер',
  'roles.doctor': 'Врач',
  'roles.electrician': 'Электрик',
  'roles.analyst': 'Аналитик',
  'roles.manager': 'Менеджер',
  'roles.programmer': 'Программист',
  'roles.default': 'Другое',

  // Vacancy status
  'status.open': 'Открыта',
  'status.archived': 'В архиве',
  'status.draft': 'Черновик',

  // Vacancy create
  'create.back': '← Вакансии',
  'create.eyebrow': 'Новая',
  'create.title': 'Откройте позицию.',
  'create.subtitle':
    'Название, бриф и взвешенные навыки. Cadence сам определит роль и направит кандидатов к нужному ревьюеру.',

  // Vacancy form
  'form.titleLabel': 'Название',
  'form.titlePlaceholder': 'Например, Senior Go Engineer',
  'form.descriptionLabel': 'Описание',
  'form.descriptionPlaceholder':
    'За что отвечает роль, ожидания, стек.',
  'form.descriptionCounter': '{count} / 4000',
  'form.roleLabel': 'Роль (необязательно)',
  'form.rolePlaceholder':
    'programmer / manager / accountant — определится автоматически',
  'form.roleHint':
    'Оставьте пустым, чтобы Cadence сам определил роль по названию и описанию.',
  'form.submit': 'Опубликовать',

  // Form validation + server errors
  'form.error.titleRequired': 'Название обязательно',
  'form.error.titleMax': 'Не длиннее 255 символов',
  'form.error.descriptionMax': 'Не длиннее 4000 символов',
  'form.error.skillsMin': 'Добавьте хотя бы один навык',
  'form.error.skillNameRequired': 'Название навыка обязательно',
  'form.error.skillNameMax': 'Не длиннее 64 символов',
  'form.error.weightRange': 'Число от 0 до 1',
  'form.error.weightMin': 'Минимум 0',
  'form.error.weightMax': 'Максимум 1',
  'form.error.invalidArg': 'Некоторые поля заполнены неверно. Проверьте форму.',
  'form.error.unauthorized': 'Сессия истекла. Войдите снова.',
  'form.error.rateLimited':
    'Слишком много запросов — попробуйте чуть позже.',
  'form.error.network': 'Не удаётся связаться с сервером.',
  'form.error.fallback': 'Не удалось создать вакансию.',

  // Skills editor
  'skills.legend': 'Навыки и веса',
  'skills.hint': 'Вес 0–1 · обязательный или желательный',
  'skills.namePlaceholder': 'Например, Go, PostgreSQL, gRPC',
  'skills.weightPlaceholder': '0,00',
  'skills.must': 'Обяз.',
  'skills.nice': 'Желат.',
  'skills.add': '+ Добавить навык',
  'skills.remove': 'Удалить',
  'skills.empty': 'Навыков пока нет.',
  'skills.allZero':
    'Все веса равны 0 — backend распределит их поровну.',
  'skills.skillNameAria': 'Название навыка {index}',
  'skills.weightAria': 'Вес навыка {index}',
  'skills.removeAria': 'Удалить навык {index}',

  // Vacancy details
  'details.skillsHeader': 'Навыки · {n}',
  'details.skillsHeaderWithMust': 'Навыки · {n} (обяз. {m})',
  'details.uploadSuccess': '✓ {name} загружен — анализ запущен.',
  'details.uploadFailure': 'Не удалось загрузить резюме.',
  'details.candidates': 'Кандидаты',
  'details.candidatesSubtitle': 'отсортированы по match score',
  'details.empty.title': 'Загрузите первое резюме сверху.',
  'details.empty.hint':
    'Cadence создаст кандидата, извлечёт профиль и запустит анализ автоматически.',
  'details.empty.badge': 'Нет кандидатов',
  'details.selectPrompt.badge': 'Выберите кандидата',
  'details.selectPrompt.text':
    'Кликните по кандидату слева, чтобы увидеть полный анализ, разбор скоринга и обоснование AI.',
  'details.error.loadVacancy': 'Не удалось загрузить вакансию.',
  'details.error.loadCandidates': 'Не удалось загрузить кандидатов.',

  // Resume uploader
  'upload.title': 'Перетащите резюме, чтобы добавить кандидата',
  'upload.subtitle':
    'PDF, DOC, DOCX или TXT до 10 МБ. Cadence создаст кандидата, извлечёт профиль и запустит анализ автоматически.',
  'upload.cta': 'Выбрать файл',
  'upload.uploading': 'Загрузка',
  'upload.tooLarge': 'Файл больше 10 МБ. Выберите PDF или DOCX поменьше.',

  // Analysis details
  'analysis.matchScore': 'Match score',
  'analysis.recommendation': 'Рекомендация HR',
  'analysis.breakdown': 'Разбор навыков',
  'analysis.matched': 'Совпали',
  'analysis.missing': 'Не хватает',
  'analysis.extra': 'Сверх',
  'analysis.profile': 'Профиль кандидата',
  'analysis.feedback': 'Фидбэк кандидату',
  'analysis.experience': 'Опыт',
  'analysis.positions': 'Должности',
  'analysis.technologies': 'Технологии',
  'analysis.education': 'Образование',
  'analysis.yearsOne': '{n} год',
  'analysis.yearsFew': '{n} года',
  'analysis.yearsMany': '{n} лет',
  'analysis.error.load': 'Не удалось загрузить анализ.',
  'analysis.downloadResume': 'Скачать резюме',
  'analysis.downloadFailed': 'Не удалось скачать резюме.',
  'analysis.deleteCandidate': 'Удалить кандидата',
  'analysis.deleteConfirm': 'Точно удалить?',
  'analysis.deleteCancel': 'Отмена',
  'analysis.deleteConfirmCta': 'Удалить',
  'analysis.deleteFailed': 'Не удалось удалить кандидата.',

  // Analysis status
  'analysis.status.queued': 'В очереди',
  'analysis.status.running': 'Идёт',
  'analysis.status.done': 'Готово',
  'analysis.status.failed': 'Ошибка',
  'analysis.status.unknown': '—',

  // Recommendation badges
  'rec.hire': 'Нанимать',
  'rec.maybe': 'Возможно',
  'rec.no': 'Нет',

  // Candidate row
  'candidate.unnamed': 'Кандидат без имени',
  'candidate.match': 'мэтч',

  // Footer
  'footer.tagline': 'Найм в спокойном ритме.',
  'footer.privacy': 'Конфиденциальность',
  'footer.terms': 'Условия',
  'footer.help': 'Поддержка',
  'footer.navAria': 'Навигация в подвале',
  'footer.copyright': '© {year} Cadence',

  // a11y
  'a11y.skipToMain': 'Перейти к содержимому',
  'a11y.openMenu': 'Открыть меню',
  'a11y.closeMenu': 'Закрыть меню',
}

const en: Dict = {
  // Common
  'common.signIn': 'Sign in',
  'common.signOut': 'Sign out',
  'common.cancel': 'Cancel',
  'common.error': 'Error',
  'common.loading': 'Loading',
  'common.back': '← Back',
  'common.create': 'Create',
  'common.save': 'Save',
  'common.search': 'Search',
  'common.email': 'Email',
  'common.password': 'Password',

  // Nav
  'nav.vacancies': 'Vacancies',

  // Theme toggle
  'theme.toLight': 'Switch to light theme',
  'theme.toDark': 'Switch to dark theme',

  // Locale toggle
  'locale.label': 'Interface language',
  'locale.ru': 'Russian',
  'locale.en': 'English',

  // Auth
  'auth.eyebrow': 'Institutional hiring',
  'auth.titleLogin': 'Welcome back to Cadence.',
  'auth.titleRegister': 'Hire on a steady cadence.',
  'auth.subtitle':
    'Vacancy intake, AI-assisted resume scoring, and decisions you can defend — on a single, quiet canvas.',
  'auth.modeLogin': 'Sign in',
  'auth.modeRegister': 'Create account',
  'auth.formTitleLogin': 'Sign in',
  'auth.formTitleRegister': 'Create your account',
  'auth.formHintLogin': 'Use your work email and password.',
  'auth.formHintRegister':
    'Eight characters minimum. Use a passphrase you can remember.',
  'auth.confirmPassword': 'Confirm password',
  'auth.passwordHintRegister': 'At least 8 characters',
  'auth.submitLogin': 'Sign in',
  'auth.submitRegister': 'Create account',

  'auth.error.emailRequired': 'Email is required',
  'auth.error.emailInvalid': 'Invalid email format',
  'auth.error.passwordShort': 'At least 8 characters',
  'auth.error.passwordLong': 'Maximum 72 characters',
  'auth.error.passwordsNoMatch': 'Passwords do not match',
  'auth.error.invalidCreds': 'Invalid email or password',
  'auth.error.exists': 'An account with this email already exists',
  'auth.error.invalidArg': 'Check the entered data',
  'auth.error.rateLimited': 'Too many attempts — wait a moment and retry',
  'auth.error.network': 'Cannot reach the server',
  'auth.error.fallbackLogin': 'Sign in failed. Try again.',
  'auth.error.fallbackRegister': 'Registration failed. Try again.',

  // Vacancies list
  'vacancies.eyebrow': 'Pipeline',
  'vacancies.title': 'Vacancies',
  'vacancies.subtitle':
    'Open positions, scored candidates, decisions you can defend.',
  'vacancies.new': 'New vacancy',
  'vacancies.searchPlaceholder': 'Search by title, description',
  'vacancies.countOne': '{n} vacancy',
  'vacancies.countFew': '{n} vacancies',
  'vacancies.countMany': '{n} vacancies',
  'vacancies.empty.filteredTitle': 'Nothing matches that filter.',
  'vacancies.empty.filteredHint':
    'Try a different role or clear the search.',
  'vacancies.empty.title': 'No vacancies yet.',
  'vacancies.empty.hint':
    'Open the first position and start collecting candidates.',
  'vacancies.empty.cta': 'Create vacancy',
  'vacancies.empty.badgeFiltered': 'No results',
  'vacancies.empty.badge': 'Empty',
  'vacancies.error.loadFailed': 'Failed to load vacancies',
  'vacancies.error.title': "We couldn't load vacancies.",

  // Roles
  'roles.all': 'All roles',
  'roles.accountant': 'Accountant',
  'roles.doctor': 'Doctor',
  'roles.electrician': 'Electrician',
  'roles.analyst': 'Analyst',
  'roles.manager': 'Manager',
  'roles.programmer': 'Programmer',
  'roles.default': 'Other',

  // Vacancy status
  'status.open': 'Open',
  'status.archived': 'Archived',
  'status.draft': 'Draft',

  // Vacancy create
  'create.back': '← Vacancies',
  'create.eyebrow': 'New',
  'create.title': 'Open a position.',
  'create.subtitle':
    'Title, brief, and weighted skills. Cadence will infer the role and route candidates to the right reviewer prompt.',

  // Vacancy form
  'form.titleLabel': 'Title',
  'form.titlePlaceholder': 'e.g. Senior Go Engineer',
  'form.descriptionLabel': 'Description',
  'form.descriptionPlaceholder':
    'What this role is responsible for, expectations, stack.',
  'form.descriptionCounter': '{count} / 4000',
  'form.roleLabel': 'Role (optional)',
  'form.rolePlaceholder':
    'programmer / manager / accountant — auto-detected if empty',
  'form.roleHint':
    'Leave empty to let Cadence infer the role from title and description.',
  'form.submit': 'Publish vacancy',

  'form.error.titleRequired': 'Title is required',
  'form.error.titleMax': 'Up to 255 characters',
  'form.error.descriptionMax': 'Up to 4000 characters',
  'form.error.skillsMin': 'Add at least one skill',
  'form.error.skillNameRequired': 'Skill name is required',
  'form.error.skillNameMax': 'Up to 64 characters',
  'form.error.weightRange': 'Number from 0 to 1',
  'form.error.weightMin': 'Min 0',
  'form.error.weightMax': 'Max 1',
  'form.error.invalidArg': 'Some fields are invalid. Check the form.',
  'form.error.unauthorized': 'Your session expired. Sign in again.',
  'form.error.rateLimited': 'Too many requests — try again in a moment.',
  'form.error.network': 'Cannot reach the server.',
  'form.error.fallback': 'Could not create the vacancy.',

  // Skills editor
  'skills.legend': 'Skills & weights',
  'skills.hint': 'Weight 0–1 · must-have or nice-to-have',
  'skills.namePlaceholder': 'e.g. Go, PostgreSQL, gRPC',
  'skills.weightPlaceholder': '0.00',
  'skills.must': 'Must',
  'skills.nice': 'Nice',
  'skills.add': '+ Add skill',
  'skills.remove': 'Remove',
  'skills.empty': 'No skills yet.',
  'skills.allZero':
    'All weights are 0 — backend will spread them equally.',
  'skills.skillNameAria': 'Skill {index} name',
  'skills.weightAria': 'Skill {index} weight',
  'skills.removeAria': 'Remove skill {index}',

  // Vacancy details
  'details.skillsHeader': 'Skills · {n}',
  'details.skillsHeaderWithMust': 'Skills · {n} ({m} must)',
  'details.uploadSuccess': '✓ {name} uploaded — analysis started.',
  'details.uploadFailure': 'Could not upload resume.',
  'details.candidates': 'Candidates',
  'details.candidatesSubtitle': 'ranked by match score',
  'details.empty.title': 'Upload the first resume above.',
  'details.empty.hint':
    'Cadence will create the candidate, extract a profile, and run analysis automatically.',
  'details.empty.badge': 'No candidates',
  'details.selectPrompt.badge': 'Select a candidate',
  'details.selectPrompt.text':
    'Click a candidate on the left to see the full analysis, score breakdown, and AI rationale.',
  'details.error.loadVacancy': 'Could not load vacancy.',
  'details.error.loadCandidates': 'Could not load candidates.',

  // Resume uploader
  'upload.title': 'Drop a resume to add a candidate',
  'upload.subtitle':
    'PDF, DOC, DOCX, or TXT up to 10 MB. Cadence will create the candidate, extract a profile, and run analysis automatically.',
  'upload.cta': 'Choose file',
  'upload.uploading': 'Uploading',
  'upload.tooLarge': 'File is over 10 MB. Pick a smaller PDF or DOCX.',

  // Analysis details
  'analysis.matchScore': 'Match score',
  'analysis.recommendation': 'HR recommendation',
  'analysis.breakdown': 'Skills breakdown',
  'analysis.matched': 'Matched',
  'analysis.missing': 'Missing',
  'analysis.extra': 'Extra',
  'analysis.profile': 'Candidate profile',
  'analysis.feedback': 'Candidate feedback',
  'analysis.experience': 'Experience',
  'analysis.positions': 'Positions',
  'analysis.technologies': 'Technologies',
  'analysis.education': 'Education',
  'analysis.yearsOne': '{n} year',
  'analysis.yearsMany': '{n} years',
  'analysis.error.load': 'Could not load analysis.',
  'analysis.downloadResume': 'Download resume',
  'analysis.downloadFailed': 'Could not download resume.',
  'analysis.deleteCandidate': 'Remove candidate',
  'analysis.deleteConfirm': 'Remove this candidate?',
  'analysis.deleteCancel': 'Cancel',
  'analysis.deleteConfirmCta': 'Remove',
  'analysis.deleteFailed': 'Could not remove candidate.',

  // Analysis status
  'analysis.status.queued': 'Queued',
  'analysis.status.running': 'Running',
  'analysis.status.done': 'Done',
  'analysis.status.failed': 'Failed',
  'analysis.status.unknown': '—',

  // Recommendation badges
  'rec.hire': 'Hire',
  'rec.maybe': 'Maybe',
  'rec.no': 'No',

  // Candidate row
  'candidate.unnamed': 'Unnamed candidate',
  'candidate.match': 'match',

  // Footer
  'footer.tagline': 'Hire on a steady cadence.',
  'footer.privacy': 'Privacy',
  'footer.terms': 'Terms',
  'footer.help': 'Help',
  'footer.navAria': 'Footer navigation',
  'footer.copyright': '© {year} Cadence',

  // a11y
  'a11y.skipToMain': 'Skip to main content',
  'a11y.openMenu': 'Open menu',
  'a11y.closeMenu': 'Close menu',
}

export const dictionaries: Record<Locale, Dict> = { ru, en }

/**
 * Russian-style pluralisation: 1, 2–4, 5–20, then mod 10.
 * Returns the suffix to append to a base key, e.g. `vacancies.count` →
 *   `vacancies.countOne | countFew | countMany`.
 * For English the mapping collapses to one/many.
 */
export function pluralKey(
  baseKey: string,
  n: number,
  locale: Locale,
): string {
  const abs = Math.abs(n)
  if (locale === 'ru') {
    const mod10 = abs % 10
    const mod100 = abs % 100
    if (mod10 === 1 && mod100 !== 11) return `${baseKey}One`
    if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14))
      return `${baseKey}Few`
    return `${baseKey}Many`
  }
  return abs === 1 ? `${baseKey}One` : `${baseKey}Many`
}
