export async function register() {
  if (process.env.NEXT_RUNTIME === "nodejs") {
    const { validateServerConfiguration } = await import("./lib/auth/runtime-config");
    try { validateServerConfiguration(); } catch (error) {
      // Next can keep its listener alive after a rejected instrumentation hook.
      // Refuse startup explicitly so Compose does not keep an unusable server.
      console.error(error instanceof Error ? error.message : "Invalid application configuration");
      process.exit(1);
    }
  }
}
