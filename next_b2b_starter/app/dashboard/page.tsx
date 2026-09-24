import Link from "next/link";
import { redirect } from "next/navigation";
import { verifyPayment } from "@/lib/actions/billing/verify-payment";
import { isPolarEnabled } from "@/lib/polar/config";
export default async function DashboardPage({ searchParams }: { searchParams: Promise<{ checkout_id?: string }> }) {
 const { checkout_id } = await searchParams;
 if (checkout_id && isPolarEnabled()) {
  const result = await verifyPayment(checkout_id);
  redirect(`/dashboard/settings?view=subscription&${result.success && result.data.HasActiveSubscription ? "payment_verified" : "payment_error"}=true`);
 }
 return <div className="mx-auto max-w-4xl px-4 py-12">
   <p className="text-sm text-gray-500">Your workspace</p><h1 className="mt-2 text-3xl font-semibold">Welcome to your dashboard</h1>
   <p className="mt-4 max-w-xl text-gray-600">Your organization is ready. Manage your account and invite your team to get started.</p>
   <div className="mt-10 grid gap-4 sm:grid-cols-2">
    <Link className="rounded-xl border p-6 hover:bg-gray-50" href="/dashboard/settings?view=profile"><h2 className="font-semibold">Account & workspace</h2><p className="mt-2 text-sm text-gray-600">Update your profile and organization details.</p></Link>
    <Link className="rounded-xl border p-6 hover:bg-gray-50" href="/dashboard/settings"><h2 className="font-semibold">Workspace settings</h2><p className="mt-2 text-sm text-gray-600">Manage your team&apos;s access and settings.</p></Link>
   </div>
 </div>;
}
