This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/app/api-reference/cli/create-next-app).

**Stack:** Node **26+** (see `.node-version`), [pnpm](https://pnpm.io/) **12** pinned by [`packageManager`](https://nodejs.org/api/packages.html#packagemanager) in `package.json`. pnpm settings (`allowBuilds`, `overrides`, etc.) live in [`pnpm-workspace.yaml`](https://pnpm.io/pnpm-workspace_yaml).

## Getting Started

From this directory:

```bash
pnpm install
pnpm dev
```

Node 26 no longer ships Corepack, so install the pnpm CLI once at the `packageManager` version, then restart the terminal:

```bash
npm install -g pnpm@12.23.0
```

Node’s installer normally puts `%AppData%\npm` on your user `PATH`. If the error persists, add that folder in **Settings → System → About → Advanced system settings → Environment Variables**. As a one-off without PATH changes, from `frontend` you can run `npx pnpm@12.23.0 dev`.

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
