"use client";

import { useId } from "react";

import { LoadingIndicator } from "@/components/ui/loading-state";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuthStore } from "@/stores/auth-store";

import { useNextEmployeeCode } from "../hooks";

type OperationalEmployeeCodeFieldProps = {
  /** Existing assigned code; skips preview fetch when set. */
  assignedCode?: string | null;
};

export function OperationalEmployeeCodeField({
  assignedCode,
}: OperationalEmployeeCodeFieldProps) {
  const fieldId = useId();
  const hintId = useId();
  const accessToken = useAuthStore((state) => state.accessToken);
  const hasAssignedCode = Boolean(assignedCode?.trim());
  const previewQuery = useNextEmployeeCode(
    !hasAssignedCode && Boolean(accessToken),
  );
  const code = hasAssignedCode
    ? assignedCode!.trim()
    : previewQuery.data?.employee_code;
  const isResolving =
    !hasAssignedCode &&
    !code &&
    (!accessToken || previewQuery.isPending || previewQuery.isFetching);

  const hint = hasAssignedCode
    ? "Assigned automatically and kept for the life of the account."
    : "Auto-assigned on save.";

  return (
    <div className="field-control field-control--readonly">
      <Label htmlFor={fieldId}>Employee code</Label>

      {isResolving ? (
        <div
          aria-busy="true"
          aria-labelledby={fieldId}
          className="field-chrome field-chrome--readonly employee-id-field-loading"
        >
          <LoadingIndicator className="employee-id-field-loading-indicator" dots={5} />
        </div>
      ) : previewQuery.isError && !code ? (
        <div className="employee-id-field-error">
          <p className="form-field-hint">Unable to load the next employee code.</p>
          <button
            className="employee-id-field-retry"
            onClick={() => void previewQuery.refetch()}
            type="button"
          >
            Try again
          </button>
        </div>
      ) : (
        <>
          <Input
            aria-describedby={hintId}
            autoComplete="off"
            id={fieldId}
            readOnly
            tabIndex={-1}
            value={code ?? ""}
          />
          <p className="form-field-hint" id={hintId}>
            {hint}
          </p>
        </>
      )}
    </div>
  );
}
