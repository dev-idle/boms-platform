import { timingSafeEqual } from "node:crypto";

import type {} from "react/experimental";
import * as React from "react";

type TaintObjectReference = (
  message: string | undefined,
  object: object,
) => void;

function readTaintObjectReference(): TaintObjectReference | undefined {
  return (
    React as unknown as {
      experimental_taintObjectReference?: TaintObjectReference;
    }
  ).experimental_taintObjectReference;
}

/**
 * Marks an object reference so it cannot be passed to Client Components.
 * Uses React's `experimental_taintObjectReference` when provided by the runtime.
 */
export function taintSensitiveObjectReference(
  object: object,
  message = "BOMS: sensitive object must not cross the RSC/client boundary.",
): void {
  const taint = readTaintObjectReference();
  if (taint) {
    taint(message, object);
    return;
  }
  if (process.env.NODE_ENV === "development") {
    console.warn(
      "[BOMS] experimental_taintObjectReference is unavailable. Use a React build that ships taint APIs to enforce reference leakage protection.",
    );
  }
}

type TaintUniqueValue = string | bigint | ArrayBufferView;

function readTaintUniqueValue(): ((
  message: string | undefined,
  lifetime: object,
  value: TaintUniqueValue,
) => void) | undefined {
  return (
    React as unknown as {
      experimental_taintUniqueValue?: (
        message: string | undefined,
        lifetime: object,
        value: TaintUniqueValue,
      ) => void;
    }
  ).experimental_taintUniqueValue;
}

/** Stable lifetime object required by `experimental_taintUniqueValue`. */
const TAINT_UNIQUE_VALUE_LIFETIME: object = Object.freeze({});

/**
 * Prevents unique secret values from being passed to Client Components.
 */
export function taintSensitiveUniqueValue(
  value: TaintUniqueValue,
  message = "BOMS: unique secret must not cross the RSC/client boundary.",
): void {
  const taint = readTaintUniqueValue();
  if (taint) {
    taint(message, TAINT_UNIQUE_VALUE_LIFETIME, value);
    return;
  }
  if (process.env.NODE_ENV === "development") {
    console.warn(
      "[BOMS] experimental_taintUniqueValue is unavailable in this React build.",
    );
  }
}

/**
 * Constant-time string comparison for secrets (UTF-8 byte length must match).
 */
export function timingSafeStringEqual(a: string, b: string): boolean {
  const aBuf = Buffer.from(a, "utf8");
  const bBuf = Buffer.from(b, "utf8");
  if (aBuf.length !== bBuf.length) {
    return false;
  }
  return timingSafeEqual(aBuf, bBuf);
}

/**
 * Redacts common secret patterns for structured logs.
 */
export function redactSecrets(input: string): string {
  return input
    .replace(/Bearer\s+[^\s]+/gi, "Bearer [REDACTED]")
    .replace(/(password|token|secret)=([^&\s]+)/gi, "$1=[REDACTED]");
}
