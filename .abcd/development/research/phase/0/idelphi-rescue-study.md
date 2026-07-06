# idelphi Rescue Study — Phase 0 Task 3

## TL;DR

- The rescue preserved the Delphi core (engine, orchestrator, facilitator, consensus, MLX infra) as recognisable units — these are the right atoms for abcd lifeboat automation (all grew post-rescue but identity was maintained)
- Surface layer was rebuilt from scratch: wizard shells, multi-provider UI, the Ingestion/ directory (7 files), and per-study state machines were all dropped
- idelphiDev grew by 34 Swift files net (121 → 155) despite shedding complexity — the gain came from JSON-config layer, analytics, convergence metrics, and richer engine telemetry
- The extraction.md 5-section template (Keep / Drop / Carry-forward ADRs / Gotchas / Files Rescued) maps cleanly onto what a lifeboat artefact must contain
- One anti-pattern discovered: `DelphiEngine.swift` grew from 589 → 1766 lines post-rescue, suggesting the "engine absorbs new behaviour" drift already started

---

## Part A: extraction.md template analysis

### Section breakdown

| # | Section heading | Role in lifeboat | Content shape |
|---|-----------------|------------------|---------------|
| 1 | What Worked Well — Keep As-Is | Positive extraction | Named file → rationale → v2 adaptation note per subsection |
| 2 | What Failed / Was Too Complex — Drop | Negative extraction | Named anti-pattern → why it failed → v2 replacement |
| 3 | Architecture Decisions That Carry Forward | Decision record | ADR ID / title → decision text → rationale |
| 4 | Key Gotchas | Tribal knowledge | Bullet list: `<symptom> — <mitigation>` |
| 5 | Files Rescued | Inventory | Table per category (code/docs/research): dir → file count → purpose |

**Subsection depth:** Each § 1 entry follows `Source: <file list>` → narrative → `Adaptation for v2:` note. This 3-part shape is the per-item contract.

**Line budget:** 140 lines total. §1 = 35 lines, §2 = 33 lines, §3 = 15 lines, §4 = 11 lines, §5 = 26 lines.

### Schema implications for Pass C composers

Pass C (principle-distiller → brief-composer → press-release-composer) must produce output that satisfies the same 5-section shape when rendered:

1. **Keep** items must cite source files, not just concepts — `code-rescuer` (Pass A) must emit file-level evidence, not principle summaries only
2. **Drop** items must name the anti-pattern and a replacement direction — `review-collator` findings plus `principle-distiller` should surface these
3. **Carry-forward ADRs** must be referenced by ID (prevents inline duplication) — `decision-archaeologist` (Pass A) owns this
4. **Gotchas** must follow `fact — mitigation` bullet form (actionable, not narrative) — `chat-distiller` (Pass B) is the right source since gotchas are tribal knowledge rarely visible in code
5. **Files Rescued** must be a table with exact directory names and file counts — `brief-composer` inventory section owns this

The `Adaptation for v2:` trailing note in §1 is load-bearing — it transforms a historical observation into a forward instruction. `press-release-composer` or `brief-composer` must emit an equivalent forward-pointer per Keep item, otherwise the lifeboat is archival only.

---

## Part B: iDelphiZero ↔ idelphiDev diff

### Tree-shape changes

**iDelphiZero structure** (under `Sources/`):
```
AppState/  Engine/  Ingestion/  Keychain/  Models/  Navigation/
Providers/  Report/  Storage/  Store/  Views/{Dashboard,ExpertLibrary,
PanelAssembly,Report,Setup,Shared,Study,Validation,Vendors,Welcome,Wizard}
```

**idelphiDev structure** (flat, no `Sources/` wrapper):
```
Analytics/  Config/  DesignSystem/  Engine/  Models/Study/
Providers/MLX/  Report/  Resources/  Services/  State/  Views/
{Analytics,Onboarding,Progress,Results,Setup,Shared}
```

**Dropped directories** (iDelphiZero only): `Ingestion/`, `Keychain/`, `Navigation/`, `Storage/`, `Store/`, `Views/Dashboard`, `Views/ExpertLibrary`, `Views/PanelAssembly`, `Views/Validation`, `Views/Vendors`, `Views/Welcome`, `Views/Wizard`

**Added directories** (idelphiDev only): `Analytics/`, `Config/` (15 JSON-config Swift files), `DesignSystem/`, `Providers/MLX/`, `Resources/`, `Services/`, `State/`, `Views/Onboarding`, `Views/Progress`, `Views/Results`

**File count:** 121 Swift files (iDelphiZero) → 155 Swift files (idelphiDev). Net +34 despite major surface removal.

### Targeted file diffs (7 files)

**1. AppState.swift** (509 → 186 lines, −323)
- **Keep:** Responsibility reduction is genuine. iDelphiZero version owned vendor bootstrap, credential restoration, provider registry, study lifecycle, and UI routing — five distinct concerns. idelphiDev version owns only MLX provider + JSON config, with study lifecycle delegated to engine.
- **Drop:** The `ActiveStudySession` / `CompletedStudySession` struct pattern nested inside AppState. These belong in the engine or a coordinator, not in global app state.

**2. iDelphiApp.swift** (363 → 40 lines, −323)
- **Keep:** Radical simplification. iDelphiZero embedded `NotificationDelegate`, `WindowCloseDelegate`, push notification handling, and window management in the app entry point. idelphiDev reduces to `@main struct` + scene declaration.
- **Drop:** `UNUserNotificationCenterDelegate` in the app entry point. Notification logic belongs in a coordinator, not `@main`.

**3. ContentView.swift** (487 → 105 lines, −382)
- **Keep:** `NavigationSplitView` with sidebar + detail columns replaces the multi-case shell state machine. Simpler mental model.
- **Drop (anti-pattern):** The `AppShell` state machine (`case .validating`, `.studyStandalone`, `.setupStandalone`, `.workspace`) was a complex routing layer inside a view. Views should not own routing state machines.

**4. DelphiEngine.swift** (589 → 1766 lines, +1177)
- **Keep:** The new `PausedRoundState`, `FacilitatorPauseReason`, and retry-on-parse-failure logic are genuine improvements — they make the engine production-ready for flaky LLM output.
- **Drop (anti-pattern):** A 1766-line actor is a single-file god object. The retry/pause state machine should be a separate `FacilitatorRetryCoordinator` type. The rescue preserved the engine correctly but post-rescue development accumulated behaviour without splitting.

**5. ConsensusDetector.swift** (110 → 141 lines, +31)
- **Keep:** Addition of `ConsensusClass` enum (strong/conditional/operational/divergent) with Speed & Metwally citation. Elevates a boolean to a four-tier taxonomy with documented thresholds.
- **Keep:** `os.log` import added for structured logging — correct production pattern.

**6. FacilitatorAgent.swift** (103 → 191 lines, +88)
- **Keep:** System prompt moved from hardcoded string literal inside the method to a pre-composed `effectiveSystemPrompt` parameter supplied by `StudyConfigBuilder`. Separates prompt construction from message assembly.
- **Drop:** The iDelphiZero version used a raw string literal for the JSON schema instruction — fragile and not testable in isolation.

**7. LLMProvider.swift** (129 → 102 lines, −27)
- **Keep:** Removal of `case cloud`, `case localServer`, `case systemBuiltIn` from `ProviderKind` and `case apiKey`, `case hostURL` from `CredentialKind`. MLX-only simplification is correct for v2 scope.
- **Keep:** Inline documentation of remaining status-machine paths is improved.

### Patterns flagged

**Keep findings:**

| # | Finding | Evidence |
|---|---------|----------|
| K1 | Engine core (DelphiEngine, RoundOrchestrator, FacilitatorAgent, ConsensusDetector) preserved with minimal changes and enriched with production features | All 4 files exist in both trees with upward line growth |
| K2 | MLX infra (MLXProvider, MLXDownloadManager, MLXModelCatalog) carried forward intact; MLXDownloadManager grew from 694 → 1339 lines adding real production hardening | Size delta confirms active improvement, not just copy |
| K3 | JSON-config layer is a genuine architectural leap — 15 new Config/*.swift files replace hardcoded layouts | `Config/` directory absent in iDelphiZero; 15 new files in idelphiDev |
| K4 | `ConsensusClass` four-tier enum replaces boolean consensus detection — measurable improvement | Added in ConsensusDetector.swift with academic citation |
| K5 | Expert.id changed from UUID to human-readable String key — correct for JSON-driven catalog | Expert.swift diff shows `id: UUID` → `id: String` |

**Drop findings (anti-patterns confirmed):**

| # | Finding | Evidence |
|---|---------|----------|
| D1 | `SetupWizardView` (29KB) and `NewStudyWizardView` (36KB) wizard pattern — step-dependent validation spread across one massive file | Files exist only in iDelphiZero; extraction.md § 2.1 confirms |
| D2 | `AppShell` state machine routing inside ContentView — view layer owning navigation logic | ContentView iDelphiZero has `switch appState.shell { case .validating... }` |
| D3 | Notification delegate and window close delegate in `@main` entry point — lifecycle concern in wrong layer | iDelphiApp.swift diff: 363 → 40 lines, both delegates dropped |
| D4 | `LLMProviderRegistry` with 5 cloud providers — credential handling, health-check, per-provider error logic mixed into single registry | File exists only in iDelphiZero; 6 provider files dropped |
| D5 | `Ingestion/` directory (7 files: GoogleDocs, OneDrive, URL, Docx) — high complexity, rarely used in practice | Directory dropped; 7 files absent in idelphiDev |

---

## Implications for abcd lifeboat automation

1. **The lifeboat artefact must have a forward-pointing `Adaptation for v2` note per Keep item.** A lifeboat that only records what existed is archival; what makes it actionable is the forward instruction. Pass C composers (`brief-composer`, `press-release-composer`) must synthesise not just what was rescued but how each item should be adapted.

2. **File inventory (§ 5) must be machine-readable.** The human-authored `extraction.md` uses a markdown table with directory / file-count / purpose columns. For abcd automation the lifeboat should emit a structured JSON inventory alongside the prose so that `abcd embark` can verify file presence without re-reading prose.

3. **Gotchas (§ 4) are tribal knowledge that transcripts rarely make explicit.** Items like "`MLXArray` is NOT Sendable — `eval()` before returning" are not discoverable from a diff alone. Pass B (transcript sampling) must be the source for gotchas; Pass C (source diff) can only surface structural anti-patterns. The two passes are complementary, not redundant.

4. **`ConsensusClass` upgrade is a post-rescue improvement, not a rescue artefact.** The automated lifeboat should capture the iDelphiZero baseline, not idelphiDev's accumulated improvements. abcd automation should snapshot source state at rescue time, not at study time.

5. **The 12 dropped directories from iDelphiZero signal scope reduction.** abcd lifeboat automation should flag directories that exist only in the predecessor as candidate Drop items, then ask the human to confirm. This is cheaper than trying to infer "was this used?" from code analysis.

---

## Anti-patterns flagged

| # | Anti-pattern | Location | Rationale |
|---|--------------|----------|-----------|
| AP1 | Wizard as single file | `SetupWizardView.swift` (29KB), `NewStudyWizardView.swift` (36KB) in iDelphiZero | Multi-step state machines embedded in views make individual steps untestable and create merge-conflict hotspots |
| AP2 | Routing state machine in ContentView | `AppShell` enum + switch in iDelphiZero `ContentView.swift` | Views should render state, not own routing logic; this pattern leaks navigation concern into the wrong layer |
| AP3 | God-object engine accumulation | `DelphiEngine.swift` grew from 589 → 1766 lines post-rescue | Engine correctly started thin but absorbed retry coordinator, pause state, and telemetry without decomposition — early sign of the same complexity that required the rescue in the first place |
| AP4 | Notification/window delegates in `@main` | `iDelphiApp.swift` iDelphiZero (363 lines) | App entry point is not a service layer; OS-level delegates belong in dedicated coordinator objects |
| AP5 | `Expert.id: UUID` as primary key for catalog items | `Expert.swift` iDelphiZero | UUID keys are opaque in JSON configs and break human-readable catalog authoring; String keys (`"cybersecurity-analyst"`) are stable, diffable, and human-editable |
