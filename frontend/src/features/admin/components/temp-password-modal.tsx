"use client";

import { useState } from "react";
import { toast } from "sonner";

import { AppDialog, AppDialogFooterActions } from "@/components/ui/app-dialog";
import { Button } from "@/components/ui/button";

import type { AdminTempPasswordPayload } from "../schemas";

type TempPasswordModalProps = {
  data: AdminTempPasswordPayload | null;
  onClose: () => void;
  open: boolean;
};

export function TempPasswordModal({
  open,
  data,
  onClose,
}: TempPasswordModalProps) {
  if (!open || !data) {
    return null;
  }

  return <TempPasswordModalContent data={data} onClose={onClose} />;
}

type TempPasswordModalContentProps = {
  data: AdminTempPasswordPayload;
  onClose: () => void;
};

function TempPasswordModalContent({
  data,
  onClose,
}: TempPasswordModalContentProps) {
  const [copied, setCopied] = useState(false);
  const [acknowledged, setAcknowledged] = useState(false);

  const tempPassword = data.temp_password;
  const userEmail = data.user.email;

  async function handleCopy(): Promise<void> {
    await navigator.clipboard.writeText(tempPassword);
    setCopied(true);
    toast.success("Temporary password copied");
  }

  return (
    <AppDialog
      closeOnBackdrop={acknowledged}
      closeOnEscape={acknowledged}
      description={`Share this password securely with ${userEmail}. It is shown only once.`}
      footer={
        <AppDialogFooterActions>
          <Button
            disabled={!acknowledged}
            onClick={onClose}
            type="button"
            variant="outline"
          >
            Close
          </Button>
        </AppDialogFooterActions>
      }
      onClose={onClose}
      open
      size="lg"
      title="Temporary password"
    >
      <div className="app-dialog-code">{tempPassword}</div>

      <div className="app-dialog-inline-actions">
        <Button onClick={() => void handleCopy()} type="button" variant="outline">
          {copied ? "Copied" : "Copy password"}
        </Button>
        <label className="app-dialog-checkbox">
          <input
            checked={acknowledged}
            className="app-dialog-checkbox-input"
            onChange={(event) => setAcknowledged(event.target.checked)}
            type="checkbox"
          />
          <span>I have copied the password</span>
        </label>
      </div>
    </AppDialog>
  );
}
