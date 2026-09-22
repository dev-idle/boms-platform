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
          const digits = typed.replace(/\D/g, "");
          let digitsBeforeCaret = typed
            .slice(0, input.selectionStart ?? typed.length)
            .replace(/\D/g, "").length;
          // What arrived is read from how many digits it added, never from the
          // event type: mobile keyboards, input methods and autofill all send a
          // whole number as one "insertText". A deletion keeps every digit left.
          // One digit is a keystroke after the fixed +84, so only a leading 0
          // (the trunk prefix) goes — a typed 84 is the Vinaphone 084 prefix.
          // More at once is a number from elsewhere and may carry +84 or 0084.
          const added = digits.length - national.length;
          let next =
            added < 0
              ? digits
              : added <= 1
                ? digits.replace(/^0+/, "")
                : nationalNumber(typed);
          // Growth past nine digits is not applied — React puts the controlled
          // value back. Shortening always is: a number stored before the rule
          // existed can be longer, and must still be editable down to size.
          if (next.length > NATIONAL_NUMBER_LENGTH && next.length > national.length) {
            return;
          }
          // Digits dropped above (a trunk 0, 84 or 00) all sat in front.
          const removed = digits.length - next.length;
          // Backspace just after a group space removes no digit; take the one before.
          const inputType = (event.nativeEvent as InputEvent).inputType;
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
