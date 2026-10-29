# Security

Participant Services is an independently developed, synthetic reference implementation. It is not an NDIA system and is not supported for production use. Use only synthetic records and keep the demonstration local. Security responses are best effort; there is no guaranteed response time or supported release series.

## Report concerns privately

Use this repository's **Security → Advisories → Report a vulnerability** option on GitHub when private vulnerability reporting is available. Include the affected commit, a minimal synthetic reproduction, expected/actual behaviour and impact. Do not include live tokens, private keys, real participant data or unrelated confidential information.

Do not disclose an unpatched vulnerability in a public issue, pull request, discussion or ordinary email address copied from Git history. If the private reporting option is unavailable, do not publish the details; wait until the maintainer has enabled a private channel. GitHub provides this feature for public repositories. The maintainer must enable it and verify the reporting button during the approved visibility-change procedure, before announcing public availability; adding this file does not enable that GitHub setting. See [GitHub’s configuration instructions](https://docs.github.com/en/code-security/how-tos/report-and-fix-vulnerabilities/configure-vulnerability-reporting/configure-for-a-repository).

## Known boundaries

- The participant API validates JWTs and enforces participant self-access. Staff/delegate policy, entitlement provisioning/revocation and production identity assurance are not implemented.
- The local issuer has an unauthenticated token endpoint and reuses a TLS key for signing. It is an isolated test fixture, not a deployable production IdP.
- The issuer's Compose health check skips certificate verification. The API Go JWKS client and documented smoke client still validate CA trust and hostname.
- Root `server.go` is a separate unauthenticated Todo scaffold with unsynchronised mutable state. Use `cmd/api` for the protected participant API; neither is approved for public hosting.
- Query/rate controls, dependency/image scanning and production traffic-draining assurance remain incomplete. AWS deployment and managed-secrets integration are proposals.

Local keys and certificates must stay under ignored `config/certs/`; runtime configuration and backup files are excluded from Docker build context. Never commit credentials, private material or real participant information. The [testing record](docs/testing.md) is evidence of specific checks, not a security certification. [Third-party components](THIRD_PARTY_NOTICES.md) retain their own licences and maintenance responsibilities.
