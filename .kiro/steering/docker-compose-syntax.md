---
inclusion: always
---

# Docker Compose Syntax Rule

## CRITICAL REQUIREMENT: Always use `docker compose` (modern syntax)

**NEVER use `docker-compose` (legacy syntax)**

This system uses the modern Docker Compose V2 syntax. You MUST always use:

✅ **CORRECT:** `docker compose`
❌ **WRONG:** `docker-compose`

## Examples

### Correct Commands:
- `docker compose up`
- `docker compose up -d`
- `docker compose down`
- `docker compose build`
- `docker compose logs`
- `docker compose exec backend bash`
- `docker compose ps`
- `docker compose restart`

### In Documentation:
Always reference the modern syntax in README files, scripts, and instructions.

### In CI/CD:
Use `docker compose` in GitHub Actions and other automation scripts.

### In Development Instructions:
When providing setup or troubleshooting instructions, always use `docker compose`.

## Rationale

The user's system uses Docker Compose V2 which uses the `docker compose` command as a plugin to the Docker CLI, rather than the standalone `docker-compose` binary. Using the wrong syntax will cause command failures.

This rule applies to:
- All documentation
- All scripts
- All CI/CD configurations  
- All verbal/written instructions
- All troubleshooting guidance