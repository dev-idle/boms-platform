"use client";

import { Button } from "@/components/ui/button";

import { AppDialog, AppDialogFooterActions } from "./app-dialog";

type ConfirmDialogProps = {
  cancelLabel?: string;
  confirmLabel?: string;
  confirmVariant?: "default" | "destructive" | "warning";
  description: string;
  isPending?: boolean;
  onCancel: () => void;
  onConfirm: () => void;
  open: boolean;
  title: string;
};

export function ConfirmDialog({
  open,
  title,
  description,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  confirmVariant = "default",
  isPending = false,
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  return (
    <AppDialog
      description={description}
      footer={
        <AppDialogFooterActions>
          <Button
            disabled={isPending}
            onClick={onCancel}
            type="button"
            variant="outline"
          >
            {cancelLabel}
          </Button>
          <Button
            disabled={isPending}
            onClick={onConfirm}
            type="button"
            variant={confirmVariant}
          >
            {isPending ? "Processing…" : confirmLabel}
          </Button>
        </AppDialogFooterActions>
      }
      isPending={isPending}
      onClose={onCancel}
      open={open}
      panelClassName="app-dialog-panel--confirm"
      size="sm"
      title={title}
    />
  );
}
