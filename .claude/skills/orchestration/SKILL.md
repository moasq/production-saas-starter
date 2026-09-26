---
name: "orchestration"
description: "Route work between the Go and Next.js projects, assign independent issue scopes, reconcile shared API contracts, and collect review evidence. Use for repository-wide coordination."
---

<!-- Generated from .agents/skills/orchestration/SKILL.md; do not edit. -->

# Repository coordination

Read `AGENTS.md`. Keep root work to routing, contracts, shared deployment/tooling
coordination and cross-project review. Domain implementation uses project-local
instructions, skills and role briefs:

| Task | Owner and entry point |
| --- | --- |
| Go endpoints, authorization, SQLC, RLS | `go-b2b-starter/AGENTS.md` and `go-b2b-starter/.agents/agents/backend-builder.md` |
| UI, accessibility and API consumers | `next_b2b_starter/AGENTS.md` and `next_b2b_starter/.agents/agents/frontend-builder.md` |
| Better Auth, cookies, memberships | `next_b2b_starter/.agents/skills/auth-integration/SKILL.md`; read-only review through `next_b2b_starter/.agents/agents/auth-reviewer.md` |
| Browser journeys | `next_b2b_starter/.agents/agents/quality-engineer.md` |
| PR review | `.agents/agents/code-reviewer.md` and `.agents/skills/pr-review/SKILL.md` |
| Tool configuration/provider connections | `.agents/skills/dev-tools/SKILL.md` and `.agents/skills/service-connections/SKILL.md` |

For requested parallel work, assign independent files/worktrees, a tested contract
and an acceptance check. Resolve shared schema ownership before dispatch. Go owns
`go-b2b-starter/apicontract/openapi.json`; the frontend consumes generated types.
A session change may require both owners, with one auth policy and separate Go
enforcement tests. Do not fork the policy into a second root skill.

When a host cannot discover a nested role, read its canonical brief or open that
project directly. Without available or authorized delegation, follow the matching
project instructions in the same session. Role files never authorize workers or
external provider operations by themselves.

Collect each changed revision, behavior checks and limitations; request a scoped
review before an already-authorized merge. Reconcile interacting branches and
verify the integrated head. Keep README and its assets unless explicitly asked to
change them. Report local fixtures, live-provider checks and deployment separately.
