# Security

Please report a suspected vulnerability privately through this repository's GitHub
security advisory flow, where available, instead of posting credentials or exploit
details in a public issue. Never include customer data or provider secrets in a
report. If private reporting is unavailable, contact the repository maintainer
privately through their GitHub profile before sharing sensitive details.

Keep `.env`, database dumps and local audit reports private. Authentication fails
closed when unconfigured. Tenant and role checks must be retained when adding
features. The supplied Docker deployment exposes Caddy only and expects HTTPS
for public use. Run dependency checks and test provider-backed flows before a
production release; a starter cannot certify an application's security.
