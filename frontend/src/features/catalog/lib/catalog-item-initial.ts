/**
 * One letter for an image-fallback tile — the spec's watermark, not an avatar.
 * Two letters read as data and make the placeholder shout over real photos.
 */
export function catalogItemInitial(name: string): string {
  const first = name.trim().charAt(0);
  return first ? first.toUpperCase() : "—";
}
