# Project Rules

## Package Manager
- Use **pnpm** as the package manager for this project
- Install dependencies: `pnpm install`
- Never use `npm`, `yarn`, or `bun`
- The lock file is `pnpm-lock.yaml`, never commit `package-lock.json`

## Frontend
- Framework: Nuxt 3
- UI Library: Vant 4
- Styling: Tailwind CSS

## Backend
- Language: Go
- Binary: `backend/bilibili-history-go` (do not commit, add to .gitignore)
