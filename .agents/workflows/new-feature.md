---
description: How to develop a new feature for NautiKube
---

# New Feature Workflow

// turbo-all

## Steps

### 1. Create a Feature Branch

```bash
git checkout main
git pull origin main
git checkout -b feat/<feature-name>
```

Branch naming conventions:
- `feat/<name>` — new features
- `fix/<name>` — bug fixes
- `refactor/<name>` — code improvements
- `docs/<name>` — documentation only
- `chore/<name>` — tooling, CI, dependencies

### 2. Write the Spec (if SDD)

For new scanners or significant features, create a spec first:

```
.agents/specs/<feature>.spec.md
```

Define: behavior, inputs/outputs, edge cases, acceptance criteria.

### 3. Develop

Write code following `.agents/CONSTITUTION.md`. Reference Skills:
- `.agents/skills/go-patterns/SKILL.md` — Go idioms
- `.agents/skills/scanner-development/SKILL.md` — Adding scanners
- `.agents/skills/kubernetes-client/SKILL.md` — K8s API usage

### 4. Run Quality Gate

```bash
make check
```

This runs in sequence: `fmt → vet → lint → vuln → test → build`.
If any step fails, fix and re-run. Do NOT commit until `make check` passes.

### 5. Commit

```bash
git add .
git commit -m "feat(scanner): add deployment scanner with severity scoring"
```

Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat(scope): description` — new feature
- `fix(scope): description` — bug fix
- `test(scope): description` — tests
- `docs: description` — documentation
- `refactor(scope): description` — code improvement
- `chore: description` — tooling/deps

### 6. Push and Create PR

```bash
git push origin feat/<feature-name>
```

GitHub Actions will run `make check` again (double validation).

### 7. Merge

After CI passes, squash merge into `main`.
Tag a version if it's a release milestone.
