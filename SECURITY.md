# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| main    | Yes       |
| < 0.1.0 | No        |

## Reporting a Vulnerability

If you discover a security vulnerability in Sino-ImgBed, please report it responsibly.

**Do not disclose security issues in public GitHub issues.**

Instead, please:

1. Email the maintainers directly at [security@yourdomain.com] (replace with actual contact)
2. Or use GitHub Private Vulnerability Reporting if enabled

Please include:

- A description of the vulnerability
- Steps to reproduce
- Potential impact assessment
- Suggested fix (if any)

We will acknowledge receipt within 48 hours and provide an initial assessment within 7 days.

## Security Best Practices for Deployment

- Change the default `JWT_SECRET` before deploying to production
- Use HTTPS for all traffic
- Keep dependencies updated
- Configure appropriate file upload size limits
- Enable proper authentication and access controls
- Regularly review API usage logs for suspicious activity
