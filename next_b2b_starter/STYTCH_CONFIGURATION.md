# Authentication configuration

See the root [setup guide](../SETUP.md) for the authoritative configuration.

The web application needs `STYTCH_PROJECT_ID`, `STYTCH_SECRET`, `STYTCH_ENV=test|live` and `APP_BASE_URL`. Keep secrets on the server. Register `${APP_BASE_URL}/authenticate` as a login and signup redirect in Stytch.

Define the `org` RBAC resource with `view` and `manage` actions. Custom `admin` grants both; `manager` and `member` grant `org:view`. Signup assigns `admin` to the first member. The backend reads the provider policy rather than trusting local role names.

Missing credentials show setup instructions, and protected APIs fail closed. Public signup creates the workspace and sends a magic link; the callback verifies it server-side and sets HttpOnly session cookies. `Secure` is enabled when `APP_BASE_URL` is HTTPS. Production must use HTTPS.
