import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";
export default defineConfig([
 ...nextVitals, ...nextTs,
 { rules: {
   // Server snapshots and URL state intentionally synchronize in effects.
   "react-hooks/set-state-in-effect": "off",
   "@typescript-eslint/no-explicit-any": "off",
 } },
 globalIgnores([".next/**", "next-env.d.ts", "out/**"]),
]);
