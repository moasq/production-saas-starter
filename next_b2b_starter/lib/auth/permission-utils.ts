/**
 * Permission Utility Functions
 * Supports wildcard permissions (e.g., "org:*" grants all org actions)
 */



/**
 * Match a granted permission against a required permission
 * Supports wildcard matching: "resource:*" matches all actions for that resource
 *
 * @example
 * matchesPermission('org:*', 'org:create') // true
 * matchesPermission('org:view', 'org:create') // false
 * matchesPermission('org:create', 'org:create') // true
 */
export function matchesPermission(
  grantedPermission: string,
  requiredPermission: string
): boolean {
  // Direct match
  if (grantedPermission === requiredPermission) {
    return true;
  }

  // Parse permission format: "resource:action"
  const grantedParts = grantedPermission.split(':');
  const requiredParts = requiredPermission.split(':');

  // Must have 2 parts
  if (grantedParts.length !== 2 || requiredParts.length !== 2) {
    return false;
  }

  const [grantedResource, grantedAction] = grantedParts;
  const [requiredResource, requiredAction] = requiredParts;

  // Resource must match
  if (grantedResource !== requiredResource) {
    return false;
  }

  // Wildcard grants all actions for the resource
  if (grantedAction === '*') {
    return true;
  }

  // Action must match
  return grantedAction === requiredAction;
}

/**
 * Check if user has a specific permission
 *
 * @param userPermissions - Array of permissions from backend
 * @param requiredPermission - Permission to check (e.g., "org:create")
 * @returns true if user has the permission (directly or via wildcard)
 */
export function hasPermission(
  userPermissions: string[],
  requiredPermission: string
): boolean {
  return userPermissions.some(grantedPermission =>
    matchesPermission(grantedPermission, requiredPermission)
  );
}

/**
 * Check if user has ANY of the specified permissions (OR logic)
 *
 * @param userPermissions - Array of permissions from backend
 * @param requiredPermissions - Array of permissions to check
 * @returns true if user has at least one of the permissions
 *
 * @example
 * hasAnyPermission(permissions, ['org:create', 'org:view'])
 * // true if user has either permission
 */
export function hasAnyPermission(
  userPermissions: string[],
  requiredPermissions: string[]
): boolean {
  if (requiredPermissions.length === 0) {
    return true;
  }

  return requiredPermissions.some(permission =>
    hasPermission(userPermissions, permission)
  );
}

/**
 * Check if user has ALL of the specified permissions (AND logic)
 *
 * @param userPermissions - Array of permissions from backend
 * @param requiredPermissions - Array of permissions to check
 * @returns true if user has all permissions
 *
 * @example
 * hasAllPermissions(permissions, ['org:view', 'org:create'])
 * // true only if user has both permissions
 */
export function hasAllPermissions(
  userPermissions: string[],
  requiredPermissions: string[]
): boolean {
  if (requiredPermissions.length === 0) {
    return true;
  }

  return requiredPermissions.every(permission =>
    hasPermission(userPermissions, permission)
  );
}

/**
 * Check if user has a specific role
 *
 * @param userRoles - Array of role names from Stytch session
 * @param role - Role to check (e.g., "admin", "manager", "member")
 * @returns true if user has the role
 */
export function hasRole(userRoles: string[], role: string): boolean {
  return userRoles.includes(role);
}

/**
 * Check if user has ANY of the specified roles
 *
 * @param userRoles - Array of role names from Stytch session
 * @param roles - Array of roles to check
 * @returns true if user has at least one of the roles
 */
export function hasAnyRole(userRoles: string[], roles: string[]): boolean {
  return roles.some(role => userRoles.includes(role));
}

/**
 * Check if user has ALL of the specified roles
 *
 * @param userRoles - Array of role names from Stytch session
 * @param roles - Array of roles to check
 * @returns true if user has all roles
 */
export function hasAllRoles(userRoles: string[], roles: string[]): boolean {
  return roles.every(role => userRoles.includes(role));
}
