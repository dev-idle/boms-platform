export type PasswordMaskStrategy = "css" | "native";

let cachedStrategy: PasswordMaskStrategy | null = null;

function supportsCssPasswordMask(): boolean {
  if (typeof CSS === "undefined" || typeof CSS.supports !== "function") {
    return false;
  }
  return CSS.supports("-webkit-text-security", "disc");
}

export function resolvePasswordMaskStrategy(): PasswordMaskStrategy {
  if (cachedStrategy !== null) {
    return cachedStrategy;
  }
  cachedStrategy = supportsCssPasswordMask() ? "css" : "native";
  return cachedStrategy;
}
