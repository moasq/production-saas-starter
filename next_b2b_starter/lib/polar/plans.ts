export interface PolarPlan {
 id: string; name: string; description: string | null;
 price: number; currency: string; interval: string;
 productId: string;
}
export function getPlanById(plans: PolarPlan[], id: string | null | undefined) { return plans.find((p) => p.id === id) ?? null; }
export function getPlanByProductId(plans: PolarPlan[], id: string | null | undefined) { return plans.find((p) => p.productId === id) ?? null; }
