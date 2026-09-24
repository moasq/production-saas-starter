export const queryKeys = {
 profile: { all: ["profile"] as const, detail: () => ["profile", "detail"] as const },
 members: {
   all: ["members"] as const, lists: () => ["members", "list"] as const,
   list: (filters: { organizationId?: string; page?: number; pageSize?: number }) => ["members", "list", filters] as const,
   detail: (id: string) => ["members", "detail", id] as const,
 },
 subscription: { all: ["subscription"] as const, status: () => ["subscription", "status"] as const },
 products: { all: ["products"] as const, list: ["products", "list"] as const },
} as const;
