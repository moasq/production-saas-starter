// components/layout/sidebar.tsx
"use client";

import { useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { usePathname } from "next/navigation";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  LayoutDashboard,
  Settings,
  Menu,
  X,
} from "lucide-react";
const mainNavigation = [
  {
    name: "Dashboard",
    href: "/dashboard",
    icon: LayoutDashboard,
  },

];

const accountNavigation = [{ name: "Settings", href: "/dashboard/settings", icon: Settings }];

interface SidebarProps {
  isCollapsed: boolean;
}

export function Sidebar({
  isCollapsed,
}: SidebarProps) {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const pathname = usePathname();

  const closeMobileMenu = () => setIsMobileMenuOpen(false);

  return (
    <>
      {/* Mobile menu button */}
      <div className="fixed left-4 top-4 z-50 lg:hidden">
        <Button
          variant="outline"
          size="icon"
          onClick={() => setIsMobileMenuOpen((open) => !open)}
          className="bg-white"
          aria-label={isMobileMenuOpen ? "Close sidebar" : "Open sidebar"}
        >
          {isMobileMenuOpen ? (
            <X className="h-4 w-4" />
          ) : (
            <Menu className="h-4 w-4" />
          )}
        </Button>
      </div>

      {/* Mobile sidebar overlay */}
      {isMobileMenuOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={closeMobileMenu}
        />
      )}

      {/* Sidebar */}
      <div
        className={cn(
          "fixed left-0 top-0 z-50 flex h-full w-64 flex-col border-r border-gray-200 bg-white transition-[transform,width] duration-200 ease-in-out lg:translate-x-0",
          isMobileMenuOpen
            ? "translate-x-0"
            : "-translate-x-full lg:translate-x-0",
          isCollapsed && "lg:w-20",
          "overflow-hidden"
        )}
      >
        {/* Logo / brand */}
        <div className="flex items-center gap-2 border-b border-gray-100 px-5 py-5">
          <div className="flex h-11 w-11 flex-none items-center justify-center">
            <Image
              src="/icon.png"
              alt="App icon"
              width={32}
              height={32}
              className="h-8 w-8 object-contain"
            />
          </div>
          <div
            className={cn(
              "max-w-[200px] overflow-hidden transition-[max-width,opacity] duration-200 ease-linear",
              isCollapsed ? "lg:max-w-0 lg:opacity-0" : "lg:opacity-100"
            )}
          >
            <div className="text-lg font-semibold tracking-tight text-gray-900">
              Your App
            </div>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto p-6">
          <div className="space-y-1">
            {mainNavigation.map((item) => {
              const isActive =
                item.href.startsWith("/dashboard/") && item.href !== "/dashboard"
                  ? pathname.startsWith(item.href)
                  : pathname === item.href;

              return (
                <Link
                  key={item.name}
                  href={item.href}
                  onClick={closeMobileMenu}
                  title={isCollapsed ? item.name : undefined}
                  className={cn(
                    "relative flex items-center overflow-hidden rounded-md px-3 py-2.5 text-sm font-medium transition-colors",
                    isCollapsed && "lg:justify-center lg:px-5",
                    isActive
                      ? "bg-gray-900 text-white"
                      : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                  )}
                  aria-label={item.name}
                >
                  <item.icon
                    className={cn(
                      "h-4 w-4 flex-none",
                      isActive ? "text-white" : "text-gray-400"
                    )}
                  />
                  <span
                    className={cn(
                      "ml-3 whitespace-nowrap transition-[margin,max-width,opacity] duration-200 ease-linear",
                      isCollapsed
                        ? "lg:ml-0 lg:max-w-0 lg:opacity-0"
                        : "lg:max-w-[160px] lg:opacity-100"
                    )}
                  >
                    {item.name}
                  </span>
                </Link>
              );
            })}
          </div>
        </nav>

        {/* Account section */}
        <div
          className={cn(
            "border-t border-gray-100 px-5 py-4 transition-[padding] duration-200",
            isCollapsed && "lg:px-3"
          )}
        >
          <div className="space-y-2.5">
            {accountNavigation.map((item) => {
              const isActive = pathname === item.href;

              return (
                <Link
                  key={item.name}
                  href={item.href}
                  onClick={closeMobileMenu}
                  title={isCollapsed ? item.name : undefined}
                  className={cn(
                    "flex h-10 items-center gap-2 rounded-md border border-transparent px-3 text-sm font-medium transition-colors",
                    isCollapsed && "lg:h-9 lg:justify-center lg:px-0",
                    isActive
                      ? "border-gray-900 bg-gray-900 text-white shadow-sm"
                      : "text-gray-600 hover:border-gray-200 hover:bg-gray-100 hover:text-gray-900"
                  )}
                >
                  <item.icon
                    className={cn(
                      "h-4 w-4 flex-none text-gray-500",
                      isActive && "text-white"
                    )}
                  />
                  <span
                    className={cn(
                      "flex-1 truncate transition-[max-width,opacity] duration-200",
                      isCollapsed
                        ? "lg:max-w-0 lg:opacity-0"
                        : "lg:max-w-[120px] lg:opacity-100"
                    )}
                  >
                    {item.name}
                  </span>
                </Link>
              );
            })}
          </div>
        </div>
      </div>
    </>
  );
}
