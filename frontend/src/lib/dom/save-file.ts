/** How long the file's object URL outlives the click that starts its download. */
const RELEASE_AFTER_MS = 30_000;

/**
 * Hands the browser a file to save: content of the given media type under
 * fileName. The object URL is released only once the download has surely
 * started — some browsers read it after click() returns, and a URL released
 * at once aborts it.
 */
export function saveFile(fileName: string, content: string, type: string): void {
  const blob = new Blob([content], { type });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = fileName;
  link.click();
  window.setTimeout(() => URL.revokeObjectURL(url), RELEASE_AFTER_MS);
}
