# Project Rules

## Package Manager
- Use **bun** as the package manager for this project
- Install dependencies: `bun install`
- Never use `npm` or `yarn`
- The lock file is `bun.lock`, never commit `package-lock.json`

## Frontend
- Framework: Nuxt 3
- UI Library: Vant 4
- Styling: Tailwind CSS

## Backend
- Language: Go
- Binary: `backend/bilibili-history-go` (do not commit, add to .gitignore)
