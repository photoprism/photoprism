---
applyTo:
- "frontend/**"
---

For frontend changes:

- Keep Vue 3 idioms; prefer options API and existing components.
- Run `make build-js` and `make test-js`; fix ESLint/Prettier before proposing changes.
- Avoid introducing global state; follow existing store patterns.
- Keep strings translatable; don’t hardcode locales.
- For performance, avoid unnecessary reactivity and re-renders; justify new deps.
