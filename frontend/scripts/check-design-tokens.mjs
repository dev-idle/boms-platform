#!/usr/bin/env node
/**
 * Design-system CI gate (Choux Matcha v3).
 *
 * Fails when src/ contains:
 *  1. Legacy Rosé palette hexes (must be zero after the v3 migration).
 *  2. The string "Rosé" (brand renamed to Choux).
 *  3. Vietnamese diacritics (UI copy is English-only).
 *  4. Hex color literals in .tsx files (colors must go through tokens).
 */
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative } from "node:path";

const SRC_ROOT = new URL("../src", import.meta.url).pathname.replace(
  /^\/([A-Za-z]:)/,
  "$1",
);

const LEGACY_HEXES = [
  // Rosé palette (v2)
  "#B9486E",
  "#A93B60",
  "#A03A5C",
  "#C2517A",
  "#86304D",
  "#F8EDF1",
  "#F7E6EB",
  "#8A3E6C",
  "#463139",
  "#3E2D36",
  "#8FA383",
  // Champagne gold (pre-brass recalibration)
  "#F3E9D8",
  "#C5A572",
  "#93753E",
];

const VIETNAMESE_DIACRITICS =
  /[ăâđêôơưĂÂĐÊÔƠƯàáảãạèéẻẽẹìíỉĩịòóỏõọùúủũụỳýỷỹỵÀÁẢÃẠÈÉẺẼẸÌÍỈĨỊÒÓỎÕỌÙÚỦŨỤỲÝỶỸỴ]/;

const HEX_LITERAL = /#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})\b/;

const TEXT_EXTENSIONS = new Set([".ts", ".tsx", ".css", ".svg", ".md"]);

function walk(dir, files = []) {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      walk(full, files);
    } else if ([...TEXT_EXTENSIONS].some((ext) => entry.endsWith(ext))) {
      files.push(full);
    }
  }
  return files;
}

const violations = [];

for (const file of walk(SRC_ROOT)) {
  const rel = relative(SRC_ROOT, file).replaceAll("\\", "/");
  const lines = readFileSync(file, "utf8").split(/\r?\n/);

  lines.forEach((line, index) => {
    const where = `src/${rel}:${index + 1}`;
    const upper = line.toUpperCase();

    for (const hex of LEGACY_HEXES) {
      if (upper.includes(hex)) {
        violations.push(`${where} legacy Rosé hex ${hex}`);
      }
    }
    if (line.includes("Ros\u00e9")) {
      violations.push(`${where} brand string "Ros\u00e9" (use "Choux")`);
    }
    if (VIETNAMESE_DIACRITICS.test(line)) {
      violations.push(`${where} Vietnamese diacritics (UI copy is English-only)`);
    }
    // Next.js viewport themeColor is browser metadata — it cannot read CSS variables.
    if (
      file.endsWith(".tsx") &&
      HEX_LITERAL.test(line) &&
      !line.includes("themeColor")
    ) {
      violations.push(`${where} hex literal in component (use tokens)`);
    }
  });
}

if (violations.length > 0) {
  console.error(`Design-token gate failed (${violations.length} violation(s)):`);
  for (const violation of violations) {
    console.error(`  ${violation}`);
  }
  process.exit(1);
}

console.log("Design-token gate passed.");
