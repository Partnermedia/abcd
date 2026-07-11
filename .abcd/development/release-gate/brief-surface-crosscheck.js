// Release-gate semantic detector — the iss-35 brief↔surface cross-check.
//
// This is a HOST-RUN agent-harness workflow, not a CI step and not a standalone
// executable: it spawns LLM checker agents, so it runs in the maintainer's agent
// harness at release time. The deterministic gates run in CI (see
// ../../../.github/workflows/release.yml); this is the semantic half. It reports
// DISCREPANCIES between the design brief's surface prose and the shipped binary's
// actual behaviour; the maintainer records the verdict as a signed, sha-keyed
// VSA-shaped receipt (see this directory's README.md and the design of record,
// ../plans/2026-07-11-iss35-semantic-release-gate.md).
//
// Invoke with args = { briefDocs: [<repo-relative paths>],
//                      surfaces: [{ name, kind, probe }] }.
export const meta = {
  name: 'iss35-brief-surface-crosscheck',
  description: 'Bidirectional brief↔surface reconciliation detector (iss-35 semantic gate)',
  phases: [
    { title: 'CheckBrief', detail: 'per brief doc: verify every surface claim against the binary/tree' },
    { title: 'CheckSurface', detail: 'per real surface: find its brief home' },
    { title: 'Merge', detail: 'dedup + classify discrepancies' },
  ],
}
const input = typeof args === 'string' ? JSON.parse(args) : args
const FINDINGS = {
  type: 'object', additionalProperties: false,
  properties: {
    item: { type: 'string' },
    discrepancies: { type: 'array', items: {
      type: 'object', additionalProperties: false,
      properties: {
        where: { type: 'string', description: 'file:line or verb/skill name' },
        claim: { type: 'string', description: 'what the record claims (or omits)' },
        reality: { type: 'string', description: 'what the binary/tree actually shows' },
        class: { type: 'string', enum: ['false-claim','undocumented-surface','fictional-layout','criterion-violation','stale-count'] },
      }, required: ['where','claim','reality','class'] } },
  }, required: ['item','discrepancies'],
}
const CTX = `Repo root: the current working directory — use repo-relative paths
throughout, never absolute local paths. Ground truth is the SHIPPED surface,
verified empirically: build the binary (make build produces bin/abcd-<goos>-<arch>)
and run it (\`abcd --help\` and \`abcd <verb> --help\`), and list commands/abcd/
and skills/. abcd currently ships ZERO skills — the whole /abcd: surface is
commands under commands/abcd/ (ahoy, capture, consult, ingest, prepare-this-repo,
docs, history, launch, memory, version); skills/ is empty or absent. The brief's
surface chapters are .abcd/development/brief/04-surfaces/*.md and
05-internals/08-skills.md. Report DISCREPANCIES ONLY — where record and reality
disagree, or one side is missing. A brief row explicitly marked staged (its
Status column is "staged") / probe-only / later-phase is NOT a discrepancy; an
unmarked claim about a surface that does not exist IS. Do not fix anything.`
const briefFindings = input.briefDocs.map(doc => () => agent(
  `${CTX}\n\nDirection A. Read ${doc} fully. Extract every checkable claim
about the shipped surface (verbs, sub-verbs, flags, skill names, counts,
file layouts, "abcd ships N ..." statements) and verify each against
reality. Return item="${doc}" and the discrepancy list.`,
  { label: `brief:${doc.split('/').pop()}`, phase: 'CheckBrief', schema: FINDINGS }))
const surfFindings = input.surfaces.map(s => () => agent(
  `${CTX}\n\nDirection B. The real surface "${s.name}" (${s.kind}) exists:
inspect it (${s.probe}). Search the brief's surface chapters for its
documented home (grep .abcd/development/brief/). If no brief row documents
it — or the brief documents it under a wrong name/shape — that is a
discrepancy. Return item="${s.name}" and the discrepancy list (empty if
properly documented).`,
  { label: `surface:${s.name}`, phase: 'CheckSurface', schema: FINDINGS }))
const all = (await parallel([...briefFindings, ...surfFindings]))
  .filter(Boolean)
log(`${all.length} checkers returned`)
const merged = []
const seen = new Set()
for (const r of all) for (const d of r.discrepancies) {
  const k = `${d.where}|${d.claim.slice(0,60)}`
  if (seen.has(k)) continue
  seen.add(k); merged.push({ item: r.item, ...d })
}
log(`${merged.length} unique discrepancies`)
return { count: merged.length, byClass: merged.reduce((m,d)=>(m[d.class]=(m[d.class]||0)+1,m),{}), discrepancies: merged }
