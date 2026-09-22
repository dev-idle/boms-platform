import type * as React from "react";

import {
  NATIONAL_NUMBER_LENGTH,
  formatNationalNumber,
  nationalNumber,
} from "@/lib/validation/phone";

import { Input } from "./input";

type PhoneInputProps = Omit<
  React.ComponentProps<"input">,
  "onChange" | "type" | "value"
> & {
  onChange: (value: string) => void;
  value: string | null | undefined;
};

/** Index just after the `count`-th digit of `text`, or 0 when `count` is 0. */
function indexAfterDigits(text: string, count: number): number {
  if (count <= 0) {
    return 0;
  }
  let seen = 0;
  for (let index = 0; index < text.length; index += 1) {
    if (/\d/.test(text[index]) && ++seen === count) {
      return index + 1;
    }
  }
  return text.length;
}

/**
 * The one phone field: a fixed `+84`, then the national number shown in groups
 * of three (`912 345 678`). Every change is reduced to that number, so a leading
 * 0, a pasted `+84 …` or `0084 …` and any separators fall away as they are
 * typed; the form only ever holds digits. Redrawing the groups would throw the
 * caret to the end, so it is put back after the same digit.
 *
 * No placeholder: digits in grey after the +84 read as a number already
 * entered, and the prefix alone says what the field takes.
 *
 * The number is nine digits and the field holds no more: a digit typed into a
 * full field is refused before it lands, and a paste that would overflow is
 * dropped rather than cut to a number that merely looks valid.
 */
export function PhoneInput({ onChange, value, ...props }: PhoneInputProps) {
  const national = value ?? "";
  return (
    <div className="phone-input">
      <span aria-hidden className="phone-input__prefix">
        +84
      </span>
      <Input
        autoComplete="tel"
        inputMode="tel"
        type="tel"
        {...props}
        onBeforeInput={(event) => {
          const input = event.currentTarget;
          const replacing = input.selectionStart !== input.selectionEnd;
          if (
            national.length >= NATIONAL_NUMBER_LENGTH &&
            !replacing &&
            /\d/.test(event.data)
          ) {
            event.preventDefault();
          }
        }}
        onChange={(event) => {
          const input = event.currentTarget;
          const typed = input.value;
          // Autofill can arrive as a plain Event with no inputType.
          const inputType = (event.nativeEvent as InputEvent).inputType ?? "";
          const digits = typed.replace(/\D/g, "");
          let digitsBeforeCaret = typed
            .slice(0, input.selectionStart ?? typed.length)
            .replace(/\D/g, "").length;
          // Keystrokes after the fixed +84 are the national number itself — a
          // typed 84 is a Vinaphone 084 prefix, not a country code. The one
          // exception is a 0 typed first, the trunk prefix, which is dropped. A
          // paste or autofill may carry +84 / 0084 and is reduced in full.
          let next: string;
          if (inputType === "insertText" || inputType.startsWith("delete")) {
            const trunkZero =
              inputType === "insertText" &&
              digitsBeforeCaret === 1 &&
              digits.startsWith("0");
            next = trunkZero ? digits.slice(1) : digits;
          } else {
            next = nationalNumber(typed);
          }
          // Growth past nine digits is not applied — React puts the controlled
          // value back. Shortening always is: a number stored before the rule
          // existed can be longer, and must still be editable down to size.
          if (next.length > NATIONAL_NUMBER_LENGTH && next.length > national.length) {
            return;
          }
          // Digits dropped above (a trunk 0, 84 or 00) all sat in front.
          const removed = digits.length - next.length;
          // Backspace just after a group space removes no digit; take the one before.
          if (inputType === "deleteContentBackward" && next === national) {
            const at = digitsBeforeCaret - removed;
            if (at > 0) {
              next = next.slice(0, at - 1) + next.slice(at);
              digitsBeforeCaret -= 1;
            }
          }
          const caret = indexAfterDigits(
            formatNationalNumber(next),
            digitsBeforeCaret - removed,
          );
          onChange(next);
          requestAnimationFrame(() => input.setSelectionRange(caret, caret));
        }}
        value={formatNationalNumber(national)}
      />
    </div>
  );
}
