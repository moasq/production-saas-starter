import { randomUUID } from "node:crypto";
import { betterAuth } from "better-auth";
import { organization } from "better-auth/plugins/organization";
import { magicLink } from "better-auth/plugins/magic-link";
import { nextCookies } from "better-auth/next-js";
import { getAuthDatabase } from "./database.ts";
import { sendAuthEmail } from "./email.ts";
import { consumeAuthLimit } from "./rate-limit.ts";
import { accessControl, roles } from "./rbac.ts";

export function createAuth() {
  const baseURL = process.env.APP_BASE_URL || "http://localhost:3000";
  const secret = process.env.BETTER_AUTH_SECRET;
  if (!secret || secret.length < 32) throw new Error("BETTER_AUTH_SECRET must contain at least 32 characters");
  return betterAuth({
    appName: "B2B SaaS Starter", baseURL, basePath: "/api/identity", secret,
    trustedOrigins: [new URL(baseURL).origin], database: getAuthDatabase(),
    session: { expiresIn: 60 * 60 * 8, disableSessionRefresh: true, cookieCache: { enabled: false } },
    advanced: { useSecureCookies: new URL(baseURL).protocol === "https:", database: { generateId: () => randomUUID() } },
    rateLimit: { enabled: true, window: 60, max: 60, customStorage: {
      consume: async (key, rule) => { const allowed = await consumeAuthLimit(`http:${key}`, rule.max, rule.window); return { allowed, retryAfter: allowed ? null : rule.window }; },
    } },
    plugins: [
      organization({
        ac: accessControl, roles, creatorRole: "admin", requireEmailVerificationOnInvitation: true,
        allowUserToCreateOrganization: (user) => user.emailVerified,
        organizationLimit: 10, membershipLimit: 100, invitationExpiresIn: 60 * 60 * 48,
        async sendInvitationEmail(data) {
          const url = new URL("/invite", baseURL); url.searchParams.set("id", data.id);
          await sendAuthEmail(data.email, `Invitation to ${data.organization.name}`,
            `You have been invited to ${data.organization.name}. Sign in with this email address and accept your invitation:\n\n${url}\n\nThis invitation expires in 48 hours.`);
        },
      }),
      magicLink({ expiresIn: 600, storeToken: "hashed", sendMagicLink: async ({ email, url }) => {
        if (!await consumeAuthLimit(`delivery:${email.toLowerCase()}`, 5, 600)) throw new Error("Too many email requests");
        await sendAuthEmail(email, "Your sign-in link", `Sign in to your workspace:\n\n${url}\n\nThis link expires in 10 minutes and can be used once. If you did not request it, ignore this email.`);
      } }),
      nextCookies(),
    ],
  });
}
let auth: ReturnType<typeof createAuth> | undefined;
export function getAuth() { return auth ??= createAuth(); }
