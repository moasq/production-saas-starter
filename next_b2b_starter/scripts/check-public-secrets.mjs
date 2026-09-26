import { readdirSync, readFileSync, statSync } from "node:fs";
import { dirname, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { execFileSync, spawnSync } from "node:child_process";

const project = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const root = resolve(project, "..");
export const canary = "starter_synthetic_secret_DO_NOT_SHIP_39_2026";
const privateName = /(?:^|_)(?:TOKEN|SECRET|PASSWORD|PRIVATE_KEY|API_KEY|DATABASE_URL)(?:_|$)/;

export function checkSource(text, path) {
  for (const name of text.match(/\bNEXT_PUBLIC_[A-Z0-9_]+\b/g) || []) {
    if (privateName.test(name.slice("NEXT_PUBLIC_".length))) {
      // Never include the line or value: source may contain a real credential.
      throw new Error(`Secret-like public variable ${name} in ${path}`);
    }
  }
}

export function checkArtifact(bytes, path) {
  if (bytes.includes(Buffer.from(canary))) throw new Error(`Synthetic secret embedded in ${path}`);
}

function files(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap(entry => {
    if (entry.isSymbolicLink() || ["node_modules", "cache"].includes(entry.name)) return [];
    const path = join(directory, entry.name);
    return entry.isDirectory() ? files(path) : entry.isFile() ? [path] : [];
  });
}

export function checkSources(base = root) {
  const tracked = execFileSync("git", ["ls-files", "-z"], { cwd: base, encoding: "utf8" }).split("\0").filter(Boolean);
  const selected = tracked.filter(path =>
    /^(?:compose[^/]*\.ya?ml|\.env\.example)$/.test(path) ||
    (path.startsWith("next_b2b_starter/") &&
      !/^next_b2b_starter\/(?:tests|scripts|\.agents|\.claude|\.codex)\//.test(path) &&
      ( /\.(?:[cm]?[jt]sx?|json|ya?ml)$/.test(path) || /\/(?:Dockerfile|\.env\.example)$/.test(path) )));
  for (const path of selected) checkSource(readFileSync(join(base, path), "utf8"), path);
  return selected.length;
}

export function checkBuild(base = project) {
  const output = join(base, ".next");
  if (!statSync(join(output, "BUILD_ID")).isFile()) throw new Error("A completed production build is required");
  const artifacts = [...files(output), ...files(join(base, "public"))];
  if (!artifacts.some(path => path.includes(`${join(".next", "static")}/`))) throw new Error("Browser build artifacts are missing");
  for (const path of artifacts) checkArtifact(readFileSync(path), relative(base, path));
  return artifacts.length;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    const [mode, image, ...extra] = process.argv.slice(2);
    if (extra.length || (image && mode !== "--image")) throw new Error("Usage: check-public-secrets.mjs --source|--build|--artifacts|--image IMAGE");
    if (mode === "--source") console.log(`Public secret names checked in ${checkSources()} source/config files.`);
    else if (mode === "--build") {
      checkSources();
      const env = { ...process.env, NEXT_TELEMETRY_DISABLED: "1", BILLING_ENABLED: "false" };
      for (const key of ["POLAR_ACCESS_TOKEN", "SMTP_PASSWORD", "BETTER_AUTH_SECRET", "AUTH_INTERNAL_SECRET", "POSTGRES_PASSWORD", "APP_DATABASE_PASSWORD", "AUTH_DATABASE_PASSWORD", "PGPASSWORD"]) env[key] = canary;
      console.log("Building production artifacts with synthetic runtime secrets...");
      const build = spawnSync("pnpm", ["build"], { cwd: project, env, stdio: ["ignore", "pipe", "pipe"], maxBuffer: 16 * 1024 * 1024 });
      if (build.error || build.status !== 0) throw new Error("Synthetic-secret production build failed");
      console.log(`Synthetic secrets absent from ${checkBuild()} production artifact files.`);
    } else if (mode === "--artifacts") console.log(`Synthetic secrets absent from ${checkBuild()} production artifact files.`);
    else if (mode === "--image" && image && /^[a-zA-Z0-9][a-zA-Z0-9_./:@-]*$/.test(image)) {
      const options = { stdio: ["ignore", "pipe", "pipe"] };
      checkArtifact(execFileSync("docker", ["image", "inspect", image], options), "image configuration");
      checkArtifact(execFileSync("docker", ["image", "history", "--no-trunc", image], options), "image build history");
      console.log("Synthetic secrets absent from image configuration and build history.");
    } else throw new Error("Usage: check-public-secrets.mjs --source|--build|--artifacts|--image IMAGE");
  } catch (error) {
    // Docker/build command errors can contain environment values; only our safe diagnostics escape.
    console.error(error.status !== undefined ? "Artifact command failed; inspect locally without publishing credentials." : error.message);
    process.exitCode = 1;
  }
}
