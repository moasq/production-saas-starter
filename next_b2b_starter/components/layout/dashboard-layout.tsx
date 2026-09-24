"use client";
import { useState } from "react";
import { Sidebar } from "./sidebar";
import { Header } from "./header";
import { cn } from "@/lib/utils";
export function DashboardLayout({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false);
  return (
    <div className="min-h-screen bg-white">
      <Sidebar isCollapsed={collapsed} />
      <Header isSidebarCollapsed={collapsed} onToggleSidebar={() => setCollapsed((value) => !value)} />
      <main className={cn("p-6 transition-[padding] duration-200", collapsed ? "lg:pl-24" : "lg:pl-64")}>
        <div className="mx-auto max-w-7xl">{children}</div>
      </main>
    </div>
  );
}
