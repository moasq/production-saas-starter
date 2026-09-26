"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { usePathname } from "next/navigation";
import { LayoutDashboard, Settings, Menu } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogTitle, DialogTrigger } from "@/components/ui/dialog";

const navigation = [
  { name: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
  { name: "Settings", href: "/dashboard/settings", icon: Settings },
];

export function Sidebar({ isCollapsed }: { isCollapsed: boolean }) {
  const [open, setOpen] = useState(false);
  const pathname = usePathname();
  useEffect(() => {
    const desktop = window.matchMedia("(min-width: 1024px)");
    const closeOnDesktop = () => { if (desktop.matches) setOpen(false); };
    desktop.addEventListener("change", closeOnDesktop);
    return () => desktop.removeEventListener("change", closeOnDesktop);
  }, []);

  function content(collapsed: boolean) {
    return <>
      <div className="flex items-center gap-2 border-b border-gray-100 px-5 py-5">
        <Image src="/icon.png" alt="" width={32} height={32} className="h-8 w-8 flex-none object-contain" />
        {!collapsed && <span className="text-lg font-semibold text-gray-900">Your App</span>}
      </div>
      <nav aria-label="Workspace" className="flex flex-1 flex-col gap-2 p-4">
        {navigation.map((item) => <Link key={item.href} href={item.href}
          onClick={() => setOpen(false)} aria-label={item.name}
          aria-current={pathname === item.href ? "page" : undefined}
          title={collapsed ? item.name : undefined}
          className={cn("flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium",
            item.name === "Settings" && "mt-auto", collapsed && "justify-center px-0",
            pathname === item.href ? "bg-gray-900 text-white" : "text-gray-600 hover:bg-gray-100 hover:text-gray-900")}>
          <item.icon className="h-4 w-4 flex-none" aria-hidden />
          {!collapsed && item.name}
        </Link>)}
      </nav>
    </>;
  }

  return <>
    <aside className={cn("fixed left-0 top-0 z-50 hidden h-full flex-col border-r border-gray-200 bg-white lg:flex", isCollapsed ? "w-20" : "w-64")}>
      {content(isCollapsed)}
    </aside>
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="icon" className="fixed left-4 top-4 z-50 bg-white lg:hidden" aria-label="Open sidebar">
          <Menu className="h-4 w-4" aria-hidden />
        </Button>
      </DialogTrigger>
      <DialogContent aria-describedby={undefined} className="left-0 top-0 flex h-full w-64 max-w-[85vw] translate-x-0 translate-y-0 flex-col gap-0 rounded-none p-0 sm:rounded-none">
        <DialogTitle className="sr-only">Workspace navigation</DialogTitle>
        {content(false)}
      </DialogContent>
    </Dialog>
  </>;
}
