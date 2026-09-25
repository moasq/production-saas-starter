import Link from "next/link";
export function AuthSetup() {
  return <main className="mx-auto max-w-xl px-6 py-24">
    <p className="text-sm font-medium text-gray-500">B2B SaaS Starter</p>
    <h1 className="mt-4 text-3xl font-semibold">Connect authentication</h1>
    <p className="mt-4 text-gray-600">The application is running. To enable sign-in and workspace creation, configure the authentication database, session secret, and SMTP delivery, then restart the application.</p>
    <p className="mt-4 text-gray-600">Follow the authentication section in SETUP.md to configure the local email inbox and production mail delivery.</p>
    <Link className="mt-8 inline-block underline" href="/">Back to home</Link>
  </main>;
}
