import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const stylesDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  "../src/styles",
);

let failed = false;

for (const fileName of fs.readdirSync(stylesDir)) {
  if (!fileName.endsWith(".css")) {
    continue;
  }

  const filePath = path.join(stylesDir, fileName);
  const lines = fs.readFileSync(filePath, "utf8").split(/\n/);
  let depth = 0;

  for (let lineIndex = 0; lineIndex < lines.length; lineIndex += 1) {
    const line = lines[lineIndex];
    const lineNumber = lineIndex + 1;

    if (depth === 1 && /^\s{4}[a-zA-Z-]+:\s/.test(line)) {
      console.error(
        `${fileName}:${lineNumber}: declaration outside a rule block`,
      );
      failed = true;
      break;
    }

    for (const char of line) {
      if (char === "{") {
        depth += 1;
      }
      if (char === "}") {
        depth -= 1;
      }
      if (depth < 0) {
        console.error(`${fileName}:${lineNumber}: unexpected closing brace`);
        failed = true;
        break;
      }
    }

    if (failed) {
      break;
    }
  }

  if (!failed && depth !== 0) {
    console.error(`${fileName}: missing ${depth} closing brace(s)`);
    failed = true;
  }
}

if (failed) {
  process.exit(1);
}

console.log("CSS brace check passed");
