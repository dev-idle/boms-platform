/** How long the file's object URL outlives the click that starts its download. */
const RELEASE_AFTER_MS = 30_000;

/**
 * Hands the browser a file to save: the value as indented JSON under fileName.
 * The object URL is released only once the download has surely started — some
 * browsers read it after click() returns, and a URL released at once aborts it.
 */
export function saveJsonFile(fileName: string, value: unknown): void {
  const blob = new Blob([JSON.stringify(value, null, 2)], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = fileName;
  link.click();
  window.setTimeout(() => URL.revokeObjectURL(url), RELEASE_AFTER_MS);
}
