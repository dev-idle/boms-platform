#!/usr/bin/env node
/**
 * Writes `src/app/icon.svg` and `apple-icon.svg` from `src/constants/brand-mark.json`.
 * Next.js serves these as favicon / apple-touch-icon.
 */
import { readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const appDir = join(root, "src/app");
const brandMark = JSON.parse(
  readFileSync(join(root, "src/constants/brand-mark.json"), "utf8"),
);

function buildAppIconSvg() {
  const paths = brandMark.paths
    .map(
      (path) =>
        `<path d="${path}" stroke="${brandMark.strokeHex}" stroke-width="${brandMark.strokeWidth}" stroke-linecap="round" stroke-linejoin="round"/>`,
    )
    .join("\n  ");

  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${brandMark.viewBox} ${brandMark.viewBox}" fill="none" shape-rendering="geometricPrecision">
  <rect width="${brandMark.viewBox}" height="${brandMark.viewBox}" rx="${brandMark.iconCornerRadius}" fill="${brandMark.iconBgHex}"/>
  ${paths}
</svg>
`;
}

const appIconSvg = buildAppIconSvg();
const iconPath = join(appDir, "icon.svg");
const appleIconPath = join(appDir, "apple-icon.svg");

writeFileSync(iconPath, appIconSvg, "utf8");
writeFileSync(appleIconPath, appIconSvg, "utf8");

console.log(`Wrote ${iconPath}`);
console.log(`Wrote ${appleIconPath}`);
