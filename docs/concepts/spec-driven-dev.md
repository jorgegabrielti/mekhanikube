# Spec-Driven Development

NautiKube is developed using a strict **Spec-Driven Development (SDD)** lifecycle. 

This means that code is never written based on vague tickets or conversational assumptions. Every piece of logic must map directly to a declared Specification.

## The `.agents/` Infrastructure

If you look at the root of the NautiKube repository, you will find an `.agents/` folder. This acts as the project's explicit rulebook.

```text
.agents/
├── CONSTITUTION.md          # 1. Project Laws
├── skills/                  # 2. How to write code here
├── specs/                   # 3. Exactly what features must do
└── workflows/               # 4. Standard operating procedures
```

### 1. Constitution

The `CONSTITUTION.md` is the highest law. It defines constraints like:
- "No global mutable state"
- "Sentinel errors must be used"
- "Table-driven tests only"

### 2. Specs

In `.agents/specs/`, we define exactly what each Scanner must detect. Before a new scanner is written, a `.spec.md` is created listing the Acceptance Criteria.

### 3. The SDD Flow

When adding a feature, developers MUST follow this exact sequence:

1. Write the **Spec**
2. Write the **Test** (mapping directly to the Spec's Acceptance Criteria)
3. Ensure the test fails
4. Write the **Implementation**
5. Ensure the test passes
6. Run the **Quality Gates** (`make check`)

By forcing this flow, NautiKube achieves nearly 100% test coverage and ensures that edge cases (like `nil` pointers on Kubernetes objects) are caught before hitting production clusters.
