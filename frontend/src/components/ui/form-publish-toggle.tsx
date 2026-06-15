"use client";

import { cn } from "@/lib/utils";

type FormPublishToggleProps = {
  checked: boolean;
  description: string;
  disabled?: boolean;
  id: string;
  label: string;
  onCheckedChange: (checked: boolean) => void;
};

/** Dashboard publish toggle — matches field chrome, replaces raw checkboxes. */
export function FormPublishToggle({
  checked,
  description,
  disabled = false,
  id,
  label,
  onCheckedChange,
}: FormPublishToggleProps) {
  return (
    <div className="form-publish-toggle">
      <div className="form-publish-toggle-copy">
        <p className="form-publish-toggle-label" id={`${id}-label`}>
          {label}
        </p>
        <p className="form-publish-toggle-desc" id={`${id}-description`}>
          {description}
        </p>
      </div>
      <button
        aria-checked={checked}
        aria-describedby={`${id}-description`}
        aria-labelledby={`${id}-label`}
        className={cn(
          "form-publish-switch",
          checked && "form-publish-switch--on",
        )}
        disabled={disabled}
        id={id}
        onClick={() => onCheckedChange(!checked)}
        role="switch"
        type="button"
      >
        <span className="sr-only">{checked ? "On" : "Off"}</span>
        <span aria-hidden className="form-publish-switch-thumb" />
      </button>
    </div>
  );
}
