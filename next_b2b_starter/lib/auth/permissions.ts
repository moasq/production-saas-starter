export const PERMISSIONS = { ORG_VIEW: "org:view", ORG_MANAGE: "org:manage" } as const;
export type Permission = (typeof PERMISSIONS)[keyof typeof PERMISSIONS];
