"use client";
import { createContext, useContext, useMemo, useState, useEffect, type ReactNode } from "react";
import type { ProfileResponseDto } from "@/lib/api/api/dto/profile.dto";
import { hasPermission, hasAnyPermission, hasAllPermissions, hasRole, hasAnyRole, hasAllRoles } from "@/lib/auth/permission-utils";
type AuthState = { profile: ProfileResponseDto | null; roles: string[]; permissions: string[] };
export interface AuthContextValue extends AuthState {
  isInitialized: boolean; isAuthenticated: boolean;
  hasPermission: (permission: string) => boolean;
  hasAnyPermission: (permissions: string[]) => boolean;
  hasAllPermissions: (permissions: string[]) => boolean;
  hasRole: (role: string) => boolean;
  hasAnyRole: (roles: string[]) => boolean;
  hasAllRoles: (roles: string[]) => boolean;
  updateAuthState: (state: Partial<AuthState>) => void;
  clearAuthState: () => void;
}
const AuthContext = createContext<AuthContextValue | null>(null);
export function AuthProvider({ children, initialProfile, initialRoles, initialPermissions }: {
  children: ReactNode; initialProfile: ProfileResponseDto | null; initialRoles: string[]; initialPermissions: string[];
}) {
  const [state, setState] = useState<AuthState>({ profile: initialProfile, roles: initialRoles, permissions: initialPermissions });
  // Refresh UI state when a server navigation re-verifies the session. No stored roles are trusted.
  useEffect(() => { setState({ profile: initialProfile, roles: initialRoles, permissions: initialPermissions }); }, [initialProfile, initialRoles, initialPermissions]);
  const value = useMemo<AuthContextValue>(() => ({
    ...state, isInitialized: true, isAuthenticated: Boolean(state.profile),
    hasPermission: (p) => hasPermission(state.permissions, p),
    hasAnyPermission: (p) => hasAnyPermission(state.permissions, p),
    hasAllPermissions: (p) => hasAllPermissions(state.permissions, p),
    hasRole: (r) => hasRole(state.roles, r), hasAnyRole: (r) => hasAnyRole(state.roles, r), hasAllRoles: (r) => hasAllRoles(state.roles, r),
    updateAuthState: (next) => setState((current) => ({ ...current, ...next })),
    clearAuthState: () => setState({ profile: null, roles: [], permissions: [] }),
  }), [state]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
export function useAuthContext() { return useContext(AuthContext); }
