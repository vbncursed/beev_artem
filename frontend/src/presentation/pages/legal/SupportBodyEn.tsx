import { H2, H3, P } from './LegalLayout'
import { SupportContactCard } from './SupportContactCard'

export function SupportBodyEn() {
  return (
    <>
      <SupportContactCard
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
