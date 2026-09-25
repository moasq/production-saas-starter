export type ActionResult<T = void> =
 | { success: true; data: T }
 | { success: false; error: string; details?: string };
export function createActionError(error: string, details?: string): ActionResult<never> { return { success: false, error, details }; }
export function createActionSuccess<T>(data: T): ActionResult<T> { return { success: true, data }; }
