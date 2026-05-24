"use client";

import { useEffect, useState } from "react";
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
  const [copied, setCopied] = useState(false);
  const [acknowledged, setAcknowledged] = useState(false);

  useEffect(() => {
    if (!open) {
      setCopied(false);
      setAcknowledged(false);
    }
  }, [open]);

  if (!open || !data) {
    return null;
  }

  const tempPassword = data.temp_password;
  const userEmail = data.user.email;

  async function handleCopy(): Promise<void> {
    await navigator.clipboard.writeText(tempPassword);
    setCopied(true);
    toast.success("Temporary password copied");
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-zinc-950/60 px-4">
      <div className="w-full max-w-lg rounded-lg border border-zinc-200 bg-white p-6 shadow-xl dark:border-zinc-800 dark:bg-zinc-900">
        <h2 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50">
          Temporary password
        </h2>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-300">
          Share this password securely with{" "}
          <span className="font-medium">{userEmail}</span>. It is shown only
          once.
        </p>

        <div className="mt-4 rounded-md border border-zinc-200 bg-zinc-50 p-3 font-mono text-sm dark:border-zinc-700 dark:bg-zinc-950">
          {tempPassword}
        </div>

        <div className="mt-4 flex items-center gap-3">
          <Button onClick={() => void handleCopy()} type="button">
            {copied ? "Copied" : "Copy password"}
          </Button>
          <label className="flex items-center gap-2 text-sm text-zinc-700 dark:text-zinc-300">
            <input
              checked={acknowledged}
              className="h-4 w-4 accent-zinc-900"
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
