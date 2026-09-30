/**
 * Callback ref for the heading of a view that replaces a form: focus moves to
 * it, so a screen reader reads the outcome and the keyboard goes on from there
 * instead of from the top of the page.
 */
export function focusOnMount(node: HTMLElement | null): void {
  node?.focus();
}
