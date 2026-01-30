# Security Policy

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| 1.x     | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability in Linkrift, please report it responsibly.

**Do NOT open a public GitHub issue for security vulnerabilities.**

### How to Report

1. Email: **security@linkrift.io**
2. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your report within **48 hours**.
- **Assessment**: We will assess the vulnerability and determine its severity within **5 business days**.
- **Fix**: Critical vulnerabilities will be patched within **7 days**. Lower-severity issues will be addressed in the next scheduled release.
- **Disclosure**: We will coordinate disclosure with you. We ask that you do not publicly disclose the vulnerability until we have released a fix.

### Severity Levels

| Severity | Description | Response Time |
|----------|-------------|---------------|
| Critical | Remote code execution, SQL injection, auth bypass | 24-48 hours |
| High     | Privilege escalation, data exposure | 3-5 days |
| Medium   | CSRF, information disclosure | Next release |
| Low      | Minor issues, hardening suggestions | Best effort |

### Recognition

We appreciate responsible disclosure. With your permission, we will:
- Credit you in the security advisory
- Add you to our Hall of Fame (if established)

### Scope

The following are in scope:
- Linkrift API server (`cmd/api`)
- Linkrift redirect service (`cmd/redirect`)
- Linkrift web dashboard (`web/`)
- Authentication and authorization flows
- Data handling and storage

The following are out of scope:
- Third-party services and dependencies (report to their maintainers)
- Social engineering attacks
- Denial of service attacks
- Issues in development/testing environments

## Security Best Practices

When contributing to Linkrift, please follow these guidelines:

- Never commit secrets, API keys, or credentials
- Use parameterized queries (sqlc handles this)
- Validate and sanitize all user input
- Use HTTPS in production
- Follow the principle of least privilege
- Keep dependencies up to date

## Security Tools

We use the following tools to maintain security:

- **gosec**: Static analysis for Go security issues
- **govulncheck**: Go vulnerability database checker
- **Trivy**: Container image scanning
- **Dependabot**: Automated dependency updates
