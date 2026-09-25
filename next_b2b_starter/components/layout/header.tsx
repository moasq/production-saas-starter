// components/layout/header.tsx
"use client";

import Link from "next/link";
import { useMemo } from "react";
import {
  ChevronRight,
  PanelLeftClose,
  PanelLeftOpen,
} from "lucide-react";
import { usePathname } from "next/navigation";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { UserMenu } from "./user-menu";

export function Header({ isSidebarCollapsed, onToggleSidebar }: {
  isSidebarCollapsed: boolean;
  onToggleSidebar: () => void;
}) {
  const pathname = usePathname();

  const breadcrumbItems = useMemo(() => {
    const segments = pathname.split("/").filter(Boolean);

    // Add Dashboard as first item if not already on dashboard
    const items: Array<{ label: string; href: string; isLast: boolean }> = [];

    if (segments.length > 0 && segments[0] !== 'dashboard') {
      items.push({
        label: 'Dashboard',
        href: '/dashboard',
        isLast: false,
      });
    }

    segments.forEach((segment, index) => {
      const href = `/${segments.slice(0, index + 1).join("/")}`;

      const label = segment
        .replace(/-/g, " ")
        .replace(/\b\w/g, (char) => char.toUpperCase());

      items.push({
        label,
        href,
        isLast: index === segments.length - 1,
      });
    });

    return items;
  }, [pathname]);

  const pageTitle = breadcrumbItems[breadcrumbItems.length - 1]?.label ?? "Overview";

  return (
    <header className="sticky top-0 z-40 border-b border-gray-200 bg-white/90 backdrop-blur supports-[backdrop-filter]:bg-white/80">
      <div className="flex">
        <div
          className={cn(
            "hidden border-r border-gray-200 transition-[width] duration-200 lg:block",
            isSidebarCollapsed ? "w-20" : "w-64"
          )}
        />

        <div className="flex-1 px-6 py-4">
          <div className="flex flex-col gap-4">
            <div className="flex items-start justify-between gap-4">
              <div className="flex items-center gap-3">
                <Button
                  variant="outline"
                  size="icon"
                  onClick={onToggleSidebar}
                  className="hidden h-9 w-9 lg:inline-flex"
                  aria-label={isSidebarCollapsed ? "Expand sidebar" : "Collapse sidebar"}
                >
                  {isSidebarCollapsed ? (
                    <PanelLeftOpen className="h-4 w-4" />
                  ) : (
                    <PanelLeftClose className="h-4 w-4" />
                  )}
                </Button>

                <span className="hidden h-8 w-px bg-gray-200 lg:block" aria-hidden="true" />

                <div>
                  <h1 className="text-lg font-semibold text-gray-900">{pageTitle}</h1>
                  <nav className="mt-1 flex flex-wrap items-center gap-1 text-sm text-gray-500">
                    {breadcrumbItems.map((item, index) => (
                      <span key={item.href} className="flex items-center gap-1">
                        {item.isLast ? (
                          <span className="font-medium text-gray-700">{item.label}</span>
                        ) : (
                          <Link
                            href={item.href}
                            className="transition-colors hover:text-gray-900"
                          >
                            {item.label}
                          </Link>
                        )}
                        {index < breadcrumbItems.length - 1 && (
                          <ChevronRight className="h-3.5 w-3.5 text-gray-300" />
                        )}
                      </span>
                    ))}
                  </nav>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <UserMenu />
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}
