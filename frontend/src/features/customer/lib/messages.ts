import { BRAND } from "@/constants/brand";
import type { Message } from "@/lib/schemas/message";

/** The counter's messages read as the bakery's: who at the counter wrote one is theirs to keep. */
export function bakeryAuthor(message: Message): string | null {
  return message.from === "staff" ? BRAND.name : null;
}
