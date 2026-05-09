/**
 * Contract-first API surface. Align field names with Go JSON tags.
 */
export type ApiErrorBody = {
  code: string;
  message: string;
  requestId?: string;
};

export class BomsApiError extends Error {
  readonly status: number;
  readonly body?: unknown;

  constructor(message: string, status: number, body?: unknown) {
    super(message);
    this.name = "BomsApiError";
    this.status = status;
    this.body = body;
  }
}

export class BomsValidationError extends Error {
  readonly issues: unknown;

  constructor(message: string, issues: unknown) {
    super(message);
    this.name = "BomsValidationError";
    this.issues = issues;
  }
}
