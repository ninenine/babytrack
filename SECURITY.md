# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

Only the latest minor release receives security updates. We recommend always running the most recent version.

## Reporting a Vulnerability

We take security vulnerabilities seriously, especially given that BabyTrack handles sensitive family and child data.

### How to Report

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, report vulnerabilities via GitHub's private security advisory feature:

1. Go to the [Security Advisories](https://github.com/ninenine/babytrack/security/advisories) page
2. Click "New draft security advisory"
3. Fill in the details of the vulnerability

Alternatively, email security concerns to the repository maintainers directly.

### What to Include

- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Suggested fix (if any)

### What to Expect

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution Target**: Critical issues within 14 days, others within 30 days

### Scope

We are interested in vulnerabilities including but not limited to:

- Authentication/authorisation bypass
- SQL injection
- Cross-site scripting (XSS)
- Cross-site request forgery (CSRF)
- Sensitive data exposure
- Insecure direct object references
- Server-side request forgery (SSRF)

### Out of Scope

- Denial of service attacks
- Social engineering
- Physical security
- Issues in dependencies (report these upstream)

### Recognition

We appreciate responsible disclosure and will acknowledge security researchers in our release notes (unless you prefer to remain anonymous).
