# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**B2B SaaS Starter** - A production-ready Next.js 16 B2B SaaS application with Stytch B2B authentication, Polar.sh billing, and comprehensive RBAC system. Built for scalability, security, and developer experience.

## Development Commands

- **Development**: `pnpm dev` (uses Turbopack)
- **Build**: `pnpm build`
- **Production**: `pnpm start`
- **Lint**: `pnpm lint`
- **Package Manager**: `pnpm` only

## Tech Stack & Architecture

- **Framework**: Next.js 16.0.10 + App Router
- **Language**: TypeScript (strict mode)
- **Styling**: Tailwind CSS
- **Components**: shadcn/ui for consistent, accessible UI components
- **State Management**: TanStack Query (React Query) for server state, Zustand for client state
- **Authentication**: Stytch B2B with magic links, session management, and RBAC
- **Billing**: Polar.sh for subscriptions, usage metering, and webhooks
- **Data Fetching**: Server Actions + TanStack Query + Repository pattern
- **Philosophy**: Production-ready, secure, maintainable, performant

## Key Principles

- **Security First**: Authentication guards, permission checks, subscription validation on every sensitive operation
- **Modern Architecture**: Server Actions for mutations, React Query for caching, Next.js 16 App Router for routing
- **Type Safety**: Strict TypeScript throughout the application with comprehensive type definitions
- **Performance**: Optimized bundle size, fast load times, efficient caching strategies
- **Maintainability**: Clear separation of concerns, consistent patterns, comprehensive documentation

## Project Structure

```
app/
├── layout.tsx                    # Root layout with providers
├── page.tsx                      # Landing page
├── auth/                         # Authentication pages
├── authenticate/                 # Magic link callback
├── signup/                       # Organization signup
├── dashboard/                    # Protected dashboard routes
│   ├── page.tsx                 # Dashboard home (redirects to settings)
│   ├── settings/                # Settings page with tabs
│   └── knowledge/               # Knowledge/chat feature
└── api/                         # API routes (minimal - 2 routes)
    ├── auth/session/refresh/    # JWT refresh endpoint
    └── billing/webhook/         # Polar webhook receiver

components/
├── ui/                          # shadcn/ui components
├── layout/                      # Layout components (header, sidebar, user menu)
├── billing/                     # Billing components (plans modal, subscription status)
├── members/                     # Member management components
└── cognitive/                   # AI chat components

lib/
├── actions/                     # Server Actions
│   ├── auth/                   # Auth actions (send magic link, consume, logout)
│   └── billing/                # Billing actions (checkout, cancel, verify payment)
├── api/                        # API client and repositories
│   └── api/
│       ├── client/             # API client with auth, retry, error handling
│       └── repositories/       # Repository pattern for Go backend
├── auth/                       # Authentication utilities
│   ├── stytch/                # Stytch B2B client setup
│   ├── constants.ts           # Cookie names, routes
│   ├── server-permissions.ts  # Permission checking
│   └── token-utils.ts         # JWT utilities
├── polar/                      # Polar billing integration
│   ├── client.ts              # Polar SDK client
│   ├── subscription.ts        # Subscription fetching
│   ├── current-subscription.ts # Subscription state resolution
│   ├── plans.ts               # Plan definitions
│   └── usage.ts               # Usage metering
├── contexts/                   # React contexts
│   └── auth-context.tsx       # Auth state management
├── hooks/                      # Custom hooks
│   ├── queries/               # TanStack Query hooks
│   └── mutations/             # TanStack Mutation hooks
├── models/                     # TypeScript type definitions
├── providers/                  # Provider components
├── stores/                     # Zustand stores
└── utils/                      # Utility functions
    └── server-action-helpers.ts # ActionResult type and helpers

docs/                           # Comprehensive documentation
├── 01-getting-started.md
├── 02-authentication.md
├── 03-permissions-and-roles.md
├── 04-payments-and-billing.md
├── 05-making-api-requests.md
├── 06-creating-pages.md
├── 07-creating-apis.md
├── 08-using-hooks.md
├── 09-adding-a-feature.md
├── 10-server-actions.md
├── 11-feature-guards.md
├── 12-subscription-patterns.md
└── API-LOGGING.md
```

## Documentation

Comprehensive guides in `docs/`:
- **01-10**: Core guides (getting started, auth, permissions, billing, APIs, hooks, etc.)
- **11-feature-guards.md**: Protecting features with auth, permission, and subscription guards
- **12-subscription-patterns.md**: Managing subscriptions, checkout, and billing with Polar.sh

## Current State

- **Production-ready** authentication with Stytch B2B (magic links, sessions, RBAC)
- **Subscription billing** with Polar.sh (checkout, webhooks, usage metering)
- **Server Actions** for secure mutations with auth/permission/subscription guards
- **TanStack Query** for optimized data fetching and caching
- **Repository pattern** for Go backend integration via API client
- **Comprehensive documentation** for all major features

## Development Guidelines

### Components
- Use shadcn/ui for UI components
- Create custom components in appropriate directories (e.g., `components/billing/`, `components/members/`)
- Always add proper TypeScript types

### State Management
- **Server State**: TanStack Query (React Query) for API data
- **Client State**: Zustand for global UI state, React built-ins for local component state
- **Auth State**: AuthContext with sessionStorage persistence

### Data Fetching
- **Queries**: Use TanStack Query hooks (e.g., `useProfileQuery()`, `useSubscriptionQuery()`)
- **Mutations**: Use TanStack Mutation hooks (e.g., `useInviteMember()`, `useUpdateProfile()`)
- **Server Actions**: For operations that need server-side logic (auth, billing, etc.)

### Authentication & Authorization
- **Server Components**: Use `getMemberSession()` and `getServerPermissions()`
- **Client Components**: Use `useAuth()` and `usePermissions()` hooks
- **Server Actions**: Always check auth, permissions, and subscription status

### Styling
- Tailwind CSS for all styling
- Follow design system patterns from shadcn/ui
- Use utility classes, avoid custom CSS

### Type Safety
- Maintain strict TypeScript throughout
- Define types in `lib/models/` directory
- Use `ActionResult<T>` type for Server Action return values

## Before Adding Any Package

1. Ask: "Does this solve a real problem better than existing solutions?"
2. Check: "Is this package well-maintained and widely adopted?"
3. Verify: "Does this align with our tech stack and architecture?"
4. Consider: "What's the bundle size impact and maintenance overhead?"

## Key Files

- **docs/** - Comprehensive feature documentation
- **STYTCH_CONFIGURATION.md** - Stytch B2B setup and configuration
- **package.json** - Project dependencies and scripts
- **tailwind.config.ts** - Tailwind configuration with design tokens
- **components.json** - shadcn/ui configuration
- **lib/utils/server-action-helpers.ts** - Server Action utilities
- **lib/auth/server-permissions.ts** - Permission system
- **lib/polar/current-subscription.ts** - Subscription state resolution

## Architecture Notes

### Authentication Flow
1. User enters email → `sendMagicLink()` Server Action
2. Stytch sends magic link email
3. User clicks link → `/authenticate` page
4. `consumeMagicLink()` Server Action validates token
5. Session cookies set (httpOnly, secure)
6. User redirected to dashboard

### Permission System
- Roles: `owner`, `admin`, `member`, `approver`
- Permissions: `org:view`, `org:manage`, `resource:view`, `resource:create`, `resource:edit`, `resource:delete`
- Check via `getServerPermissions()` server-side or `usePermissions()` client-side

### Subscription Guards
- Check subscription status with `resolveCurrentSubscription()` or `useSubscriptionQuery()`
- Gate premium features behind active subscription checks
- Handle subscription states: active, inactive, no customer, authentication required

### Server Actions Pattern
```typescript
'use server';

import { getMemberSession } from '@/lib/auth/stytch/server';
import { getServerPermissions } from '@/lib/auth/server-permissions';
import { createActionError, createActionSuccess } from '@/lib/utils/server-action-helpers';

export async function myAction() {
  // 1. Auth check
  const session = await getMemberSession();
  if (!session?.session_jwt) {
    return createActionError('Authentication required.');
  }

  // 2. Permission check
  const permissions = await getServerPermissions(session);
  if (!permissions.canDoSomething) {
    return createActionError('Insufficient permissions.');
  }

  // 3. Business logic
  // ...

  return createActionSuccess(data);
}
```

The goal is to build production-ready B2B SaaS applications with enterprise-grade authentication, billing, and permission systems using modern tools and patterns.
