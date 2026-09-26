import "server-only";
import { billingEnabled } from "./environment";
export function isPolarEnabled(): boolean { return billingEnabled(); }
