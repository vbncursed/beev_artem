import type { Dict } from '../types'

export const en: Dict = {
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

  // Admin dashboard
  'admin.eyebrow': 'Control',
  'admin.title': 'Platform overview.',
  'admin.subtitle':
    'Aggregate stats, HR accounts, role management.',
  'admin.stats.users': 'Users',
  'admin.stats.usersSub': '{admins} admins',
  'admin.stats.vacancies': 'Vacancies',
  'admin.stats.candidates': 'Candidates',
  'admin.stats.analyses': 'Analyses',
  'admin.stats.analysesSub': '{done} done · {failed} failed',
  'admin.users.eyebrow': 'HR users',
  'admin.users.subtitle': 'every account on the platform',
  'admin.users.empty': 'No users yet.',
  'admin.users.vacancies': 'vacancies',
  'admin.users.candidates': 'candidates',
  'admin.users.role.admin': 'Admin',
  'admin.users.role.user': 'User',
  'admin.users.promote': 'Make admin',
  'admin.users.demote': 'Demote',
  'admin.errors.stats': 'Failed to load stats.',
  'admin.errors.users': 'Failed to load users.',
  'admin.errors.role': 'Failed to change role.',
  'nav.admin': 'Control',

  // Legal pages
  'legal.lastUpdated': 'Updated: {date}',
  'legal.eyebrow.privacy': 'Policy',
  'legal.eyebrow.terms': 'Agreement',
  'legal.eyebrow.support': 'Support',
  'legal.privacy.title': 'Privacy.',
  'legal.privacy.subtitle':
    'How Cadence handles data from HR users and candidates.',
  'legal.terms.title': 'Terms of service.',
  'legal.terms.subtitle':
    'Rules for using the service — what is allowed, what is not, what we promise.',
  'legal.support.title': 'Support.',
  'legal.support.subtitle':
    'Questions about the product, setup help, bug reports.',
}

