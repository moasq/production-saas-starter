"use client";
import { useAuthContext } from "@/lib/contexts/auth-context";
export function usePermissions() {
  const context = useAuthContext();
  if (!context) throw new Error("AuthProvider is required");
  return context;
}
