import "server-only";
export function isPolarEnabled(): boolean { return process.env.BILLING_ENABLED === "true"; }
