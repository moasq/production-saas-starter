"use client";
import { Sidebar } from "./sidebar";
import { Header } from "./header";
import { useSidebarStore } from "@/lib/stores/sidebar-store";
import { usePermissions } from "@/lib/hooks/use-permissions";
import { cn } from "@/lib/utils";
export function DashboardLayout({ children }: { children: React.ReactNode }) {
 const permissions = usePermissions();
 const collapsed = useSidebarStore((s) => s.isCollapsed);
 return <div className="min-h-screen bg-white"><Sidebar permissions={permissions} /><Header />
 <main className={cn("p-6 transition-[padding] duration-200", collapsed ? "lg:pl-24" : "lg:pl-64")}>
 <div className="mx-auto max-w-7xl">{children}</div></main></div>;
}
