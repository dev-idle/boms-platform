"use client";

import { useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";

import type { CreateOperationalResponse } from "../schemas";

type TempPasswordModalProps = {
  open: boolean;
  data: CreateOperationalResponse | null;
  onClose: () => void;
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
  data: CreateOperationalResponse;
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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-ink/40 px-4">
      <div className="w-full max-w-lg rounded-modal border border-border bg-surface p-6 shadow-hover">
        <h2 className="text-lg font-medium text-ink">
          Temporary password
        </h2>
        <p className="mt-2 text-sm text-ink-2">
          Share this password securely with{" "}
          <span className="font-medium">{userEmail}</span>. It is shown only
          once.
        </p>

        <div className="mt-4 rounded-input border border-border bg-surface-alt p-3 font-mono text-sm">
          {tempPassword}
        </div>

        <div className="mt-4 flex items-center gap-3">
          <Button onClick={() => void handleCopy()} type="button">
            {copied ? "Copied" : "Copy password"}
          </Button>
          <label className="flex items-center gap-2 text-sm text-ink">
            <input
              checked={acknowledged}
              className="h-4 w-4 accent-rose-500"
              onChange={(event) => setAcknowledged(event.target.checked)}
              type="checkbox"
            />
            I have copied the password
          </label>
        </div>

        <div className="mt-6 flex justify-end">
          <Button
            disabled={!acknowledged}
            onClick={onClose}
            type="button"
            variant="outline"
          >
            Close
          </Button>
        </div>
      </div>
    </div>
  );
}
