# Skill Registry

**Delegator use only.** Any agent that launches sub-agents reads this registry to resolve compact rules, then injects them directly into sub-agent prompts. Sub-agents do NOT read this registry or individual SKILL.md files.

## User Skills

| Trigger | Skill | Path |
|---------|-------|------|
| When creating a pull request, opening a PR, or preparing changes for review | branch-pr | /home/metadandy/.config/opencode/skills/branch-pr/SKILL.md |
| When creating a GitHub issue, reporting a bug, or requesting a feature | issue-creation | /home/metadandy/.config/opencode/skills/issue-creation/SKILL.md |
| When user says "judgment day", "judgment-day", "review adversarial", "dual review", "doble review", "juzgar", "que lo juzguen" | judgment-day | /home/metadandy/.config/opencode/skills/judgment-day/SKILL.md |
| When writing Go tests, using teatest, or adding test coverage | go-testing | /home/metadandy/.config/opencode/skills/go-testing/SKILL.md |
| When user asks to create a new skill, add agent instructions, or document patterns for AI | skill-creator | /home/metadandy/.config/opencode/skills/skill-creator/SKILL.md |

## Compact Rules

Pre-digested rules per skill. Delegators copy matching blocks into sub-agent prompts as `## Project Standards (auto-resolved)`.

### branch-pr
- Every PR MUST link an approved issue (status:approved label)
- Every PR MUST have exactly one `type:*` label
- Branch naming: `^(feat|fix|chore|docs|style|refactor|perf|test|build|ci|revert)\/[a-z0-9._-]+$`
- PR body MUST contain: Linked Issue (Closes #N), PR Type (one checkbox), Summary (1-3 bullets), Changes Table, Test Plan, Contributor Checklist
- Conventional commits required: `^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\([a-z0-9\._-]+\))?!?: .+`
- No `Co-Authored-By` trailers in commits
- Automated checks must pass: PR Validation (issue reference, approved status, type label) + Shellcheck

### issue-creation
- Blank issues are disabled — MUST use template (bug_report.yml or feature_request.yml)
- Every issue gets `status:needs-review` automatically on creation
- A maintainer MUST add `status:approved` before any PR can be opened
- Bug reports need: Pre-flight Checks, Bug Description, Steps to Reproduce, Expected/Actual Behavior, OS, Agent/Client, Shell
- Feature requests need: Pre-flight Checks, Problem Description, Proposed Solution, Affected Area
- Questions go to Discussions, not issues
- Duplicate check required before creating

### judgment-day
- Launch TWO independent blind judge sub-agents in parallel (delegate, async) — NEVER review code yourself
- Classify WARNINGs: (real) = normal user can trigger → fix required; (theoretical) = contrived scenario → report as INFO
- Convergence: Round 1 → fix confirmed issues → re-judge; After Round 2+ → only re-judge if confirmed CRITICALs remain
- APPROVED criteria: 0 confirmed CRITICALs + 0 confirmed real WARNINGs (theoretical warnings may remain)
- Blocking: MUST NOT declare APPROVED until judges pass; MUST NOT push/commit until re-judgment completes
- After 2 fix iterations, ASK user before continuing — never escalate automatically

### go-testing
- Use table-driven tests for multiple test cases (standard Go pattern)
- Pure functions → table-driven; side effects → mock dependencies; returns error → test both success/error
- TUI testing: test Model.Update() directly for state changes, use teatest.NewTestModel() for full flows
- Golden file testing: compare output against saved ".golden" files, use `-update` flag to regenerate
- Test files live next to code (`archivo_test.go`), use testify/assert and testify/mock
- Commands: `go test ./...`, `go test -v ./...`, `go test -cover ./...`

### skill-creator
- Create skill when pattern is used repeatedly, project conventions differ from generic, or complex workflows need steps
- Skill structure: `SKILL.md` (required), `assets/` (optional - templates, schemas), `references/` (optional - local docs)
- Naming: `{technology}` for generic, `{project}-{component}` for project-specific, `{action}-{target}` for workflows
- Frontmatter required: name, description (include Trigger), license (Apache-2.0), metadata.author (gentleman-programming), metadata.version
- DO: start with critical patterns, use tables for decision trees, keep examples minimal
- DON'T: add Keywords section, duplicate existing docs, use web URLs in references (use local paths)

## Project Conventions

| File | Path | Notes |
|------|------|-------|
| AGENTS.md | /home/metadandy/Projects/go-fiber-skeleton/AGENTS.md | Index — references below |
| Framework docs | https://gofiber.io | Fiber v3 (not v2 as README says) |
| Modular pattern | src/core/{module}/ | Handler → Service → Repo, DI in src/container.go |
| Error handling | api_error package | Services return `*api_error.Error`, handlers propagate to global middleware |
| Testing | Bruno + go test | testify/assert + testify/mock, tests live next to code |
| Database | PostgreSQL + Goose | No AutoMigrate, migrations in migration/ |
| Auth | src/core/auth/ | JWT via middleware.Jwt, RBAC in role/permission modules |
| Mail | src/service/mail/ | Polymorphic: Mailpit (dev) vs Resend (prod) |
| Gotcha | README.md | Outdated: mentions src/modules/ and fiber v2 — trust the code |
