This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/app/api-reference/cli/create-next-app).

**Stack:** Node **24+** (see `.node-version`), [pnpm](https://pnpm.io/) **11** via [`packageManager`](https://nodejs.org/api/packages.html#packagemanager) in `package.json` — turn on [Corepack](https://nodejs.org/api/corepack.html) once (`corepack enable`) so `pnpm` matches that pin. pnpm 11 settings (`allowBuilds`, etc.) live in [`pnpm-workspace.yaml`](https://pnpm.io/pnpm-workspace_yaml).

## Getting Started

From this directory:

```bash
corepack enable   # optional if pnpm is already on PATH at 11.0.9
pnpm install
pnpm dev
```

If `corepack enable` fails on Windows (permissions under `Program Files`), use `npx pnpm@11.0.9 install` or install pnpm from [pnpm.io/installation](https://pnpm.io/installation).

If the shell says **`pnpm` is not recognized** (Windows), install the CLI once (same major as `packageManager`), then restart the terminal:

```bash
npm install -g pnpm@11.0.9
```

Node’s installer normally puts `%AppData%\npm` on your user `PATH`. If the error persists, add that folder in **Settings → System → About → Advanced system settings → Environment Variables**. As a one-off without PATH changes, from `frontend` you can run `npx pnpm@11.0.9 dev`.

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load [Geist](https://vercel.com/font), a new font family for Vercel.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
