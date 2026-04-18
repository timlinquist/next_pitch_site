---
name: credential-scan
description: Scan commits for leaked credentials, secrets, API keys, and tokens before pushing
user_invocable: true
---

# Credential Scan

Scan all commits on the current branch (relative to main) for accidentally committed credentials, secrets, or sensitive files.

## Steps

### 1. Check for sensitive filenames

Look for files added or modified in the branch that match suspicious patterns:

```sh
git diff main...HEAD --diff-filter=ACMR --name-only | grep -iE '\.(env|pem|key|p12|pfx|jks|keystore)$|credentials|secret|\.env\.' || true
```

If any are found, **report them as failures** — these files should almost never be committed.

### 2. Search diff content for credential values

Search the actual diff content for hardcoded secrets. Look for patterns like:

- API keys / tokens (long alphanumeric strings assigned to variables like `key`, `token`, `secret`, `password`, `credential`, `api_key`, `apiKey`)
- AWS keys (`AKIA...`)
- Private keys (`-----BEGIN.*PRIVATE KEY-----`)
- Connection strings with passwords (`://user:pass@`)
- Bearer tokens with hardcoded values
- Base64-encoded secrets in config

```sh
git diff main...HEAD -p | grep -inE '(AKIA[A-Z0-9]{16}|-----BEGIN.*(PRIVATE|RSA|EC|DSA).*KEY-----|://[^/]+:[^@]+@|sk_live_|sk_test_|rk_live_|rk_test_|ghp_[A-Za-z0-9]{36}|glpat-[A-Za-z0-9\-]{20}|xox[bpras]-[A-Za-z0-9\-]+)' || true
```

**Exclude** lines that are:
- Reading from environment variables (`os.Getenv`, `process.env`, `import.meta.env`)
- Variable/field names or type definitions (e.g. `paypalToken string`)
- Auth SDK calls (`getAccessTokenSilently`, `Bearer ${token}`)
- Comments describing what env vars are needed

### 3. Report results

- If **no issues found**: report "Credential scan passed — no secrets detected in branch commits."
- If **issues found**: list each finding with the filename and line, and **block the push** by telling the user what needs to be fixed.
