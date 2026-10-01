# Project Rules

## Package Manager
- Use **pnpm 11.22.0** as the package manager for this project
- Dependencies live in `frontend/`: run `pnpm install` from that directory
- Never use `npm`, `yarn`, or `bun`
- The lock file is `frontend/pnpm-lock.yaml`, never commit `package-lock.json`

## Frontend
- Framework: Nuxt 3
- UI Library: Vant 4
- Styling: Tailwind CSS

## Backend
- Language: Go
- Binary: `backend/bilibili-history-go` (do not commit, add to .gitignore)
