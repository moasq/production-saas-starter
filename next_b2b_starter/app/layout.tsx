import "./globals.css";
import type { Metadata } from "next";
import { Toaster } from "sonner";
import { AuthProvider } from "@/lib/contexts/auth-context";
import { authBootstrap } from "@/lib/auth/bootstrap";
import { QueryProvider } from "@/lib/providers/query-provider";
export const metadata: Metadata = {
  title: { default: "B2B SaaS Starter", template: "%s | B2B SaaS Starter" },
  description: "A simple workspace for your team, with organization accounts and optional billing.",
};
export default async function RootLayout({ children }: { children: React.ReactNode }) {
  const bootstrap = await authBootstrap();
  return <html lang="en"><body className="antialiased font-sans">
    <AuthProvider initialProfile={bootstrap.profile} initialRoles={bootstrap.roles} initialPermissions={bootstrap.permissions}>
      <QueryProvider>{children}<Toaster position="top-right" richColors /></QueryProvider>
    </AuthProvider>
  </body></html>;
}
