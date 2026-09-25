import { DashboardLayout } from "@/components/layout/dashboard-layout";
import { requireMemberSession } from "@/lib/auth/server";
export default async function Layout({ children }: { children: React.ReactNode }) {
 await requireMemberSession({ returnTo: "/dashboard" });
 return <DashboardLayout>{children}</DashboardLayout>;
}
