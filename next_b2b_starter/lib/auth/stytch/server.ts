import Stytch, { envs as stytchEnvs } from "stytch";
import type { B2BSessionsAuthenticateResponse } from "stytch";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { cache } from "react";

import { buildLoginUrl } from "@/lib/auth/stytch";
import { SESSION_COOKIE_NAME, SESSION_JWT_COOKIE_NAME } from "@/lib/auth/constants";

/**
 * Decode JWT token and extract claims without verification
 * This is safe because Stytch has already verified the token
 */
export function decodeJWT(token: string): Record<string, any> | null {
  try {
    // JWT format: header.payload.signature
    const parts = token.split(".");
    if (parts.length !== 3) {
      return null;
    }

    // Decode the payload (second part)
    const payload = parts[1];
    const decodedPayload = Buffer.from(payload, "base64url").toString("utf-8");
    return JSON.parse(decodedPayload);
  } catch {
    return null;
  }
}

/**
 * Extract roles from JWT token claims
 * Stytch B2B stores roles in namespaced claim: https://stytch.com/session.roles
 */
export function extractRolesFromJWT(token: string): string[] {
  const claims = decodeJWT(token);
  if (!claims) return [];

  // Stytch uses namespaced claims for session data
  const stytchSession = claims["https://stytch.com/session"] as { roles?: unknown[] } | undefined;
  if (stytchSession && Array.isArray(stytchSession.roles)) {
    return stytchSession.roles.filter(
      (r: unknown): r is string => typeof r === "string"
    );
  }

  // Fallback: check top-level roles (some configurations)
  if (Array.isArray(claims.roles)) {
    return claims.roles.filter((r): r is string => typeof r === "string");
  }

  return [];
}

type RequireSessionOptions = {
  returnTo?: string;
};

type StytchClient = InstanceType<typeof Stytch.B2BClient>;

let client: StytchClient | null = null;
let organizationIdPromise: Promise<string[]> | null = null;

function requiredEnv(name: string, value: string | undefined): string {
  if (!value) {
    throw new Error(
      `Missing ${name} environment variable for Stytch configuration.`
    );
  }
  return value;
}

export function getStytchB2BClient(): StytchClient {
  if (client) return client;

  const projectId = requiredEnv(
    "STYTCH_PROJECT_ID",
    process.env.STYTCH_PROJECT_ID
  );
  const secret = requiredEnv("STYTCH_SECRET", process.env.STYTCH_SECRET);
  const projectEnv =
    process.env.STYTCH_PROJECT_ENV ||
    process.env.NEXT_PUBLIC_STYTCH_PROJECT_ENV ||
    "test";

  client = new Stytch.B2BClient({
    project_id: projectId,
    secret,
    env: projectEnv === "live" ? stytchEnvs.live : stytchEnvs.test,
  });

  return client;
}

function parseAllowedOrganizationIdsEnv(): string[] {
  const raw = process.env.STYTCH_ALLOWED_ORGANIZATION_IDS;
  if (!raw) return [];

  return raw
    .split(",")
    .map((value) => value.trim())
    .filter((value) => value.length > 0);
}

async function loadOrganizationIdsFromStytch(): Promise<string[]> {
  if (organizationIdPromise) {
    return organizationIdPromise;
  }

  organizationIdPromise = (async () => {
    const discoveredIds = new Set<string>();
    const stytch = getStytchB2BClient();
    let cursor: string | undefined;

    do {
      const payload: { cursor?: string; limit?: number } = { limit: 100 };
      if (cursor) {
        payload.cursor = cursor;
      }

      const response = (await stytch.organizations.search(payload)) as {
        organizations: Array<{ organization_id: string }>;
        results_metadata?: { next_cursor?: string | null };
      };

      response.organizations.forEach((org) => {
        if (org?.organization_id) {
          discoveredIds.add(org.organization_id);
        }
      });

      cursor = response.results_metadata?.next_cursor ?? undefined;
    } while (cursor);

    return Array.from(discoveredIds);
  })().catch((error) => {
    organizationIdPromise = null;
    throw error;
  });

  return organizationIdPromise;
}

export async function getOrganizationIdsForMemberSearch(): Promise<string[]> {
  const configuredIds = parseAllowedOrganizationIdsEnv();
  if (configuredIds.length > 0) {
    return configuredIds;
  }

  const discoveredIds = await loadOrganizationIdsFromStytch();
  if (discoveredIds.length === 0) {
    throw new Error(
      "Unable to determine Stytch organization IDs. Provide STYTCH_ALLOWED_ORGANIZATION_IDS or ensure at least one organization exists."
    );
  }

  return discoveredIds;
}

export const getMemberSession = cache(
  async (): Promise<B2BSessionsAuthenticateResponse | null> => {
    const cookieStore = await cookies();
    const sessionToken = cookieStore.get(SESSION_COOKIE_NAME)?.value;
    const sessionJwt = cookieStore.get(SESSION_JWT_COOKIE_NAME)?.value;

    if (!sessionToken && !sessionJwt) {
      return null;
    }

    try {
      const client = getStytchB2BClient();

      let session: B2BSessionsAuthenticateResponse;

      if (sessionJwt) {
        session = (await client.sessions.authenticateJwt({
          session_jwt: sessionJwt,
        } as any)) as B2BSessionsAuthenticateResponse;
      } else if (sessionToken) {
        session = await client.sessions.authenticate({
          session_token: sessionToken,
        });
      } else {
        return null;
      }

      return session;
    } catch {
      // Session verification failed (expected for expired/invalid sessions)
      return null;
    }
  }
);

export async function requireMemberSession(
  options?: RequireSessionOptions
): Promise<B2BSessionsAuthenticateResponse> {
  const session = await getMemberSession();

  if (!session) {
    const redirectTarget = buildLoginUrl({
      returnTo: options?.returnTo,
    });
    redirect(redirectTarget);
  }

  return session;
}
