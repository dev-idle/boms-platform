import type { UseFormReturn } from "react-hook-form";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError, ApiErrorCode } from "@/lib/errors/api-error";

import { applyApiFormFieldErrors } from "./form-errors";
import { PHONE_TAKEN_MESSAGE } from "./phone";

const toastError = vi.hoisted(() => vi.fn());
vi.mock("sonner", () => ({ toast: { error: toastError } }));

type ProfileFields = { full_name: string; phone: string };

function fakeForm() {
  const setError = vi.fn();
  return { form: { setError } as unknown as UseFormReturn<ProfileFields>, setError };
}

const phoneTaken = new ApiError(409, {
  code: ApiErrorCode.PhoneExists,
  message: "Phone number is already in use",
});

describe("applyApiFormFieldErrors", () => {
  beforeEach(() => {
    toastError.mockClear();
  });

  it("puts phone_exists on the phone field, not in a toast", () => {
    const { form, setError } = fakeForm();

    applyApiFormFieldErrors(form, phoneTaken, ["full_name", "phone"], "Failed");

    expect(setError).toHaveBeenCalledWith("phone", { message: PHONE_TAKEN_MESSAGE });
    expect(toastError).not.toHaveBeenCalled();
  });

  it("falls back to the server message when the form has no phone field", () => {
    const { form, setError } = fakeForm();

    applyApiFormFieldErrors(form, phoneTaken, ["full_name"], "Failed");

    expect(setError).not.toHaveBeenCalled();
    expect(toastError).toHaveBeenCalledWith("Phone number is already in use");
  });

  it("maps a backend vn_phone tag to the format message", () => {
    const { form, setError } = fakeForm();
    const invalid = new ApiError(400, {
      code: ApiErrorCode.Validation,
      message: "Validation failed",
      details: { phone: "vn_phone" },
    });

    applyApiFormFieldErrors(form, invalid, ["phone"], "Failed");

    expect(setError).toHaveBeenCalledWith("phone", {
      message: "Enter a Vietnam phone number, for example 0912 345 678",
    });
  });
});
