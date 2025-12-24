# Server Actions

This guide shows how to create and use Next.js Server Actions for mutations and data operations.

## What are Server Actions?

Server Actions are asynchronous functions that run on the server. They allow you to perform mutations and operations directly from your components without creating API routes.

**Benefits:**
- Simplified code - no need for separate API routes
- Better TypeScript integration
- Automatic request deduplication
- Progressive enhancement support
- Works with React Server Components

## When to Use Server Actions

**Use Server Actions for:**
- Form submissions (login, logout, create, update, delete)
- Data mutations (creating, updating, deleting records)
- Operations that require authentication/authorization
- Actions that need server-side validation

**Don't use Server Actions for:**
- Webhooks from third-party services (use API routes)
- OAuth callbacks (use API routes)
- Public APIs that external services call
- Complex file uploads with progress tracking

## File Structure

Server Actions live in `lib/actions/`:

```
lib/actions/
├── auth/
│   ├── send-magic-link.ts
│   └── logout.ts
└── billing/
    ├── create-checkout.ts
    └── cancel-subscription.ts
```

## Creating a Server Action

### Basic Structure

```typescript
// lib/actions/auth/logout.ts
'use server'

import { redirect } from 'next/navigation';
import { cookies } from 'next/headers';
import {
  createActionSuccess,
  createActionError,
  type ActionResult
} from '@/lib/utils/server-action-helpers';

export async function logout(): Promise<ActionResult<void>> {
  try {
    // Clear session cookie
    cookies().delete('stytch_session_token');

    // Redirect to login
    redirect('/auth');
  } catch (error: any) {
    console.error('[Logout] Error:', error);
    return createActionError(
      'Failed to logout',
      process.env.NODE_ENV === 'development' ? error.message : undefined
    );
  }
}
```

### With Authentication

```typescript
// lib/actions/billing/cancel-subscription.ts
'use server'

import {
  requireActionSessionWithPermissions,
  createActionSuccess,
  createActionError,
  requirePermission,
  type ActionResult
} from '@/lib/utils/server-action-helpers';
import { getPolarClient } from '@/lib/polar/client';

export async function cancelSubscription(): Promise<ActionResult<{ subscriptionId: string }>> {
  try {
    // Require authentication and permissions
    const { session, permissions } = await requireActionSessionWithPermissions();

    // Check permissions
    const permError = requirePermission(
      permissions,
      (p) => p.canManageSubscriptions,
      'You do not have permission to cancel subscriptions'
    );
    if (permError) return permError;

    // Get organization ID
    const orgId = permissions.profile?.organization?.organization_id;
    if (!orgId) {
      return createActionError('Organization not found');
    }

    // Call Polar API
    const client = getPolarClient();
    const result = await client.subscriptions.cancel({
      organizationId: orgId
    });

    return createActionSuccess({
      subscriptionId: result.id
    });
  } catch (error: any) {
    console.error('[Cancel Subscription] Error:', error);
    return createActionError(
      'Failed to cancel subscription',
      process.env.NODE_ENV === 'development' ? error.message : undefined
    );
  }
}
```

### With Form Data

```typescript
// lib/actions/auth/send-magic-link.ts
'use server'

import {
  createActionSuccess,
  createActionError,
  withErrorHandling,
  type ActionResult
} from '@/lib/utils/server-action-helpers';
import { getStytchB2BClient } from '@/lib/auth/stytch/server';

export async function sendMagicLink(
  email: string
): Promise<ActionResult<{ message: string }>> {
  return withErrorHandling(async () => {
    // Validate input
    if (!email || typeof email !== 'string') {
      return createActionError('Email address is required');
    }

    const client = getStytchB2BClient();

    // Send magic link
    await client.magicLinks.email.loginOrSignup({
      email_address: email.toLowerCase(),
      login_redirect_url: process.env.NEXT_PUBLIC_APP_BASE_URL + '/authenticate'
    });

    return createActionSuccess({
      message: 'If an account exists with that email, a magic link has been sent.'
    });
  });
}
```

## Using Server Actions in Components

### In Client Components with useTransition

```typescript
'use client'

import { useState, useTransition } from 'react';
import { sendMagicLink } from '@/lib/actions/auth/send-magic-link';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export function MagicLinkForm() {
  const [email, setEmail] = useState('');
  const [isPending, startTransition] = useTransition();
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    startTransition(async () => {
      const result = await sendMagicLink(email);

      if (result.success) {
        setSuccess(true);
      } else {
        setError(result.error);
      }
    });
  };

  return (
    <form onSubmit={handleSubmit}>
      <Input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Enter your email"
        disabled={isPending}
      />

      {error && <p className="text-red-500">{error}</p>}
      {success && <p className="text-green-500">Magic link sent!</p>}

      <Button type="submit" disabled={isPending}>
        {isPending ? 'Sending...' : 'Send Magic Link'}
      </Button>
    </form>
  );
}
```

### With React Query (for mutations)

```typescript
'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { cancelSubscription } from '@/lib/actions/billing/cancel-subscription';
import { Button } from '@/components/ui/button';

export function CancelSubscriptionButton() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: cancelSubscription,
    onSuccess: (result) => {
      if (result.success) {
        // Invalidate subscription queries to refetch
        queryClient.invalidateQueries({ queryKey: ['subscription'] });
        alert('Subscription cancelled successfully');
      } else {
        alert(result.error);
      }
    }
  });

  return (
    <Button
      onClick={() => mutation.mutate()}
      disabled={mutation.isPending}
    >
      {mutation.isPending ? 'Cancelling...' : 'Cancel Subscription'}
    </Button>
  );
}
```

### With Form Actions (progressive enhancement)

```typescript
// lib/actions/auth/send-magic-link.ts
'use server'

export async function sendMagicLinkFormAction(formData: FormData) {
  const email = formData.get('email') as string;
  return sendMagicLink(email);
}
```

```typescript
'use client'

import { useFormState, useFormStatus } from 'react-dom';
import { sendMagicLinkFormAction } from '@/lib/actions/auth/send-magic-link';

function SubmitButton() {
  const { pending } = useFormStatus();
  return (
    <button type="submit" disabled={pending}>
      {pending ? 'Sending...' : 'Send Magic Link'}
    </button>
  );
}

export function MagicLinkForm() {
  const [state, formAction] = useFormState(sendMagicLinkFormAction, null);

  return (
    <form action={formAction}>
      <input
        type="email"
        name="email"
        placeholder="Enter your email"
        required
      />

      {state && !state.success && (
        <p className="text-red-500">{state.error}</p>
      )}
      {state?.success && (
        <p className="text-green-500">Magic link sent!</p>
      )}

      <SubmitButton />
    </form>
  );
}
```

## Authentication Helpers

Use the helpers from `lib/utils/server-action-helpers.ts`:

### Optional Authentication

```typescript
const session = await getActionSession();

if (!session) {
  return createActionError('Not authenticated');
}
```

### Required Authentication

```typescript
const session = await requireActionSession();
// Throws error if not authenticated - this code only runs if authenticated
```

### With Permissions

```typescript
const { session, permissions } = await requireActionSessionWithPermissions();

const permError = requirePermission(
  permissions,
  (p) => p.canCreateInvoices,
  'You cannot create invoices'
);
if (permError) return permError;
```

## Error Handling

### Standard Pattern

```typescript
export async function myAction(): Promise<ActionResult<MyData>> {
  try {
    // Your logic here
    return createActionSuccess(data);
  } catch (error: any) {
    console.error('[My Action] Error:', error);
    return createActionError(
      'User-friendly error message',
      process.env.NODE_ENV === 'development' ? error.message : undefined
    );
  }
}
```

### With Error Wrapper

```typescript
export async function myAction(): Promise<ActionResult<MyData>> {
  return withErrorHandling(async () => {
    // Your logic here
    return createActionSuccess(data);
  });
}
```

## Redirects in Server Actions

Use Next.js `redirect()` for navigation:

```typescript
import { redirect } from 'next/navigation';

export async function logout() {
  // Clear session
  cookies().delete('stytch_session_token');

  // Redirect to login
  redirect('/auth');
}
```

## Revalidation

Revalidate cached data after mutations:

```typescript
import { revalidatePath, revalidateTag } from 'next/cache';

export async function updateProfile(data: ProfileData) {
  // Update profile
  await profileRepository.update(data);

  // Revalidate the page
  revalidatePath('/dashboard/settings');

  // Or revalidate by tag
  revalidateTag('profile');

  return createActionSuccess({ updated: true });
}
```

## Best Practices

1. **Always use `'use server'` directive** at the top of Server Action files
2. **Return `ActionResult<T>`** for consistent error handling
3. **Validate all inputs** - never trust client data
4. **Check permissions** - re-validate on the server
5. **Log errors** with descriptive messages for debugging
6. **Use TypeScript** for type safety
7. **Keep actions focused** - one action per operation
8. **Handle errors gracefully** - return user-friendly messages

## Common Patterns

### Create Operation

```typescript
export async function createVendor(
  data: CreateVendorInput
): Promise<ActionResult<{ id: string }>> {
  const { session, permissions } = await requireActionSessionWithPermissions();

  const permError = requirePermission(
    permissions,
    (p) => p.canCreateVendors
  );
  if (permError) return permError;

  const vendor = await vendorRepository.create(data, session.session_jwt);

  revalidatePath('/dashboard/vendors');

  return createActionSuccess({ id: vendor.id });
}
```

### Update Operation

```typescript
export async function updateVendor(
  id: string,
  data: UpdateVendorInput
): Promise<ActionResult<void>> {
  const { session } = await requireActionSessionWithPermissions();

  await vendorRepository.update(id, data, session.session_jwt);

  revalidatePath('/dashboard/vendors');
  revalidatePath(`/dashboard/vendors/${id}`);

  return createActionSuccessEmpty();
}
```

### Delete Operation

```typescript
export async function deleteVendor(
  id: string
): Promise<ActionResult<void>> {
  const { session, permissions } = await requireActionSessionWithPermissions();

  const permError = requirePermission(
    permissions,
    (p) => p.canDeleteVendors
  );
  if (permError) return permError;

  await vendorRepository.delete(id, session.session_jwt);

  revalidatePath('/dashboard/vendors');

  return createActionSuccessEmpty();
}
```

## Comparison: API Routes vs Server Actions

| Feature | API Routes | Server Actions |
|---------|-----------|----------------|
| Use case | Webhooks, external APIs | Mutations, form submissions |
| Location | `app/api/` | `lib/actions/` |
| Directive | None | `'use server'` |
| Return type | `NextResponse` | `ActionResult<T>` |
| Authentication | Manual session check | Use helper functions |
| Type safety | Manual typing | Full TypeScript support |
| Form integration | Manual fetch | Native form actions |
| Revalidation | Manual | Built-in with `revalidatePath` |

## Testing Server Actions

```typescript
import { sendMagicLink } from '@/lib/actions/auth/send-magic-link';

describe('sendMagicLink', () => {
  it('should send magic link for valid email', async () => {
    const result = await sendMagicLink('test@example.com');

    expect(result.success).toBe(true);
  });

  it('should return error for invalid email', async () => {
    const result = await sendMagicLink('');

    expect(result.success).toBe(false);
    expect(result.error).toContain('required');
  });
});
```

## Next Steps

👉 **Learn about**: [Adding a Feature](./09-adding-a-feature.md)
👉 **Learn about**: [Creating APIs](./07-creating-apis.md) (for webhooks)
