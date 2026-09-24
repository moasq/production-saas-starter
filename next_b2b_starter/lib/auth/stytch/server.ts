import Stytch, { envs as stytchEnvs } from "stytch";
export interface VerifiedSession { session_jwt: string; }
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { cache } from "react";

import { buildLoginUrl } from "@/lib/auth/stytch";
import { SESSION_COOKIE_NAME, SESSION_JWT_COOKIE_NAME } from "@/lib/auth/constants";

type RequireSessionOptions = {
  returnTo?: string;
};

type StytchClient = InstanceType<typeof Stytch.B2BClient>;

let client: StytchClient | null = null;


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
    process.env.STYTCH_ENV ||
    "test";

  if (projectEnv !== "test" && projectEnv !== "live") throw new Error("STYTCH_ENV must be test or live");
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

export const getMemberSession = cache(async (): Promise<VerifiedSession | null> => {
 const store = await cookies();
 const sessionJwt = store.get(SESSION_JWT_COOKIE_NAME)?.value;
 const sessionToken = store.get(SESSION_COOKIE_NAME)?.value;
 if (!sessionJwt && !sessionToken) return null;
 try {
  const client = getStytchB2BClient();
  if (sessionJwt) {
   try {
    const verified = await client.sessions.authenticateJwt({ session_jwt: sessionJwt });
    return { session_jwt: verified.session_jwt };
   } catch { /* A valid opaque session can renew an expired JWT below. */ }
  }
  if (!sessionToken) return null;
  const verified = await client.sessions.authenticate({ session_token: sessionToken });
  return { session_jwt: verified.session_jwt };
 } catch { return null; }
});

export async function requireMemberSession(
  options?: RequireSessionOptions
): Promise<VerifiedSession> {
  const session = await getMemberSession();

  if (!session) {
    const redirectTarget = buildLoginUrl({
      returnTo: options?.returnTo,
    });
    redirect(redirectTarget);
  }

  return session;
}
