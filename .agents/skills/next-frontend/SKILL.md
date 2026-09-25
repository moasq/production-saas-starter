---
name: "next-frontend"
description: "This skill applies when building a Next.js page, changing dashboard or member UI, fixing browser API state, or reviewing accessibility and server-only boundaries in this starter."
---

# Next.js frontend workflow

Read `AGENTS.md`, then the route and its data repository under
`next_b2b_starter/lib/api/`. Reuse `components/ui/`, the current layout, and the
tokens in `app/globals.css` and `tailwind.config.ts`; preserve the existing design
unless a redesign is requested. Do not copy a different project's Tailwind version.

1. Identify the user journey, backend response contract, and permission needed for
   each action. Get backend-owned contract changes agreed before inventing fields.
2. Keep presentation separate from API/provider calls. Prefer server components;
   use client components for interaction. Keep auth/provider modules server-only,
   and use the existing HTTP-only session and same-origin API paths. For session
   changes, read `.agents/skills/auth-integration/SKILL.md`.
3. Specify loading, empty, success, validation error, unavailable, unauthorized,
   and unconfigured states where applicable. Preserve response wrappers, 204
   handling, and failed invitation delivery; a successful request may still report
   delivery failure. Do not retry mutations automatically.
4. Prefer existing dependencies and accessible controls. Verify labels, keyboard
   order, focus after dialogs, visible focus, and status announcements. Inspect any
   copied component for unexpected network calls, scripts, or HTML injection.
5. Add a behavior regression for changed request/response or state logic. Use the
   existing `next_b2b_starter/tests/` conventions. Run `pnpm verify` from
   `next_b2b_starter/`; after dependency changes also run
   `pnpm audit --prod --audit-level=high` and commit the generated lockfile.

For visible changes, exercise the actual production build at 390px and 1440px.
Cover `/`, `/auth`, `/signup`, and the changed dashboard/settings journey, including
keyboard use and the relevant failure states. Test only supported themes; this
starter currently has a light UI. Record revision, route, viewport, state, and
capture time with screenshots. Check rendered response headers using `curl -I`;
do not infer browser behavior or security headers from source alone.

Report visible changes, API contract assumptions, checks performed, and missing
browser/provider evidence. A static screenshot does not prove signup, invitation,
checkout, or portal completion.
