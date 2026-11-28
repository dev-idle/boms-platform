"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useId, type KeyboardEvent } from "react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { Form, FormControl, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { MESSAGE_MAX_LENGTH, messageInputSchemaOf, type MessageInput } from "@/lib/schemas/message";

type MessageComposerProps = {
  /** Names the box for assistive tech; it has no visible label. */
  label: string;
  placeholder: string;
  pending: boolean;
  /** The most characters it takes; a message to the counter by default. */
  maxLength?: number;
  /** Sends the message; `sent` clears the box once it is delivered. */
  onSend: (input: MessageInput, sent: () => void) => void;
};

/** Writes a message: Enter sends it, Shift + Enter breaks the line. */
export function MessageComposer({
  label,
  placeholder,
  pending,
  maxLength = MESSAGE_MAX_LENGTH,
  onSend,
}: MessageComposerProps) {
  const form = useForm<MessageInput>({
    resolver: zodResolver(messageInputSchemaOf(maxLength)),
    defaultValues: { body: "" },
  });
  const hintId = useId();
  const send = form.handleSubmit((input) => onSend(input, () => form.reset()));

  // A touch keyboard has no Shift + Enter, so there Enter breaks the line.
  function sendOnEnter(event: KeyboardEvent<HTMLTextAreaElement>): void {
    if (
      event.key === "Enter" &&
      !event.shiftKey &&
      !event.nativeEvent.isComposing &&
      window.matchMedia("(pointer: fine)").matches
    ) {
      event.preventDefault();
      if (!pending) {
        void send();
      }
    }
  }

  return (
    <Form {...form}>
      <form className="message-composer" method="post" noValidate onSubmit={(event) => void send(event)}>
        <FormField
          control={form.control}
          name="body"
          render={({ field }) => (
            <FormItem>
              <FormControl aria-describedby={hintId}>
                <textarea
                  aria-label={label}
                  className="message-composer__input"
                  placeholder={placeholder}
                  rows={1}
                  onKeyDown={sendOnEnter}
                  {...field}
                />
              </FormControl>
              <FormMessage className="message-composer__error" />
            </FormItem>
          )}
        />
        <Button aria-busy={pending || undefined} className="message-composer__send" disabled={pending} type="submit">
          {pending ? "Sending…" : "Send"}
        </Button>
        <p className="message-composer__hint" id={hintId}>
          <span>Enter</span> to send · <span>Shift + Enter</span> for a new line
        </p>
      </form>
    </Form>
  );
}
