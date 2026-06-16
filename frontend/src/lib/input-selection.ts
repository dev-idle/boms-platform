export type TextSelection = Readonly<{ end: number; start: number }>;

export function clampSelection(
  selection: TextSelection,
  valueLength: number,
): TextSelection {
  const start = Math.min(Math.max(selection.start, 0), valueLength);
  const end = Math.min(Math.max(selection.end, 0), valueLength);
  return { start, end };
}

function readLiveInputSelection(input: HTMLInputElement): TextSelection | null {
  const start = input.selectionStart;
  const end = input.selectionEnd;
  if (start === null || end === null) {
    return null;
  }
  return clampSelection({ start, end }, input.value.length);
}

/** Password inputs may report 0/0 while the caret is elsewhere. */
export function isUntrustedPasswordSelectionRead(
  inputType: string,
  valueLength: number,
  live: TextSelection,
): boolean {
  return (
    inputType === "password" &&
    live.start === 0 &&
    live.end === 0 &&
    valueLength > 0
  );
}

export function rememberTrustedSelectionRead(
  inputType: string,
  valueLength: number,
  live: TextSelection | null,
  previous: TextSelection,
): TextSelection {
  if (!live) {
    return previous;
  }
  if (isUntrustedPasswordSelectionRead(inputType, valueLength, live)) {
    return previous;
  }
  return live;
}

export function rememberTrustedInputSelection(
  input: HTMLInputElement,
  previous: TextSelection,
): TextSelection {
  return rememberTrustedSelectionRead(
    input.type,
    input.value.length,
    readLiveInputSelection(input),
    previous,
  );
}

export function restoreInputSelection(
  input: HTMLInputElement,
  selection: TextSelection,
): TextSelection {
  const clamped = clampSelection(selection, input.value.length);
  input.focus({ preventScroll: true });
  try {
    input.setSelectionRange(clamped.start, clamped.end);
  } catch {
    // Input may be disabled or not yet focusable.
  }
  return clamped;
}
