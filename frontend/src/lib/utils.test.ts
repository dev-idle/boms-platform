import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";

import { describe, expect, it } from "vitest";

import { cn } from "./utils";

const STYLES_DIR = path.resolve(__dirname, "../styles");

/** Every `.text-*` class the stylesheets define: the project's type styles. */
function typeStyleClasses(): string[] {
  const found = new Set<string>();
  for (const file of readdirSync(STYLES_DIR).filter((name) => name.endsWith(".css"))) {
    const css = readFileSync(path.join(STYLES_DIR, file), "utf8");
    for (const match of css.matchAll(/\.(text-[a-z0-9-]+)\s*[{,]/g)) {
      found.add(match[1]);
    }
  }
  return [...found];
}

describe("cn", () => {
  it("keeps a type style next to a color utility", () => {
    expect(cn("text-form-label", "text-error")).toBe("text-form-label text-error");
    expect(cn("text-price text-ink")).toBe("text-price text-ink");
  });

  it("knows every type style the stylesheets define", () => {
    const styles = typeStyleClasses();
    expect(styles.length).toBeGreaterThan(0);
    for (const style of styles) {
      // A style missing from TYPE_STYLES in utils.ts is dropped here.
      expect(cn(style, "text-error"), style).toBe(`${style} text-error`);
    }
  });

  it("still resolves conflicting Tailwind utilities", () => {
    expect(cn("text-sm", "text-lg")).toBe("text-lg");
    expect(cn("text-muted", "text-error")).toBe("text-error");
    expect(cn("text-caption", "text-caption")).toBe("text-caption");
  });
});
