"use client";

import * as SelectPrimitive from "@radix-ui/react-select";
import * as React from "react";

import { cn } from "@/lib/utils";

type ParsedSelectOption = {
  disabled?: boolean;
  label: string;
  value: string;
};

type SelectMenuSide = "top" | "bottom";

export type SelectProps = Omit<
  React.SelectHTMLAttributes<HTMLSelectElement>,
  "children" | "onChange"
> & {
  children: React.ReactNode;
  onChange?: React.ChangeEventHandler<HTMLSelectElement>;
};

function parseSelectOptions(children: React.ReactNode): ParsedSelectOption[] {
  const options: ParsedSelectOption[] = [];

  React.Children.forEach(children, (child) => {
    if (!React.isValidElement(child) || child.type !== "option") {
      return;
    }

    const optionProps = child.props as React.OptionHTMLAttributes<HTMLOptionElement>;
    const rawValue = optionProps.value;
    const value =
      rawValue === undefined || rawValue === null ? "" : String(rawValue);
    const label =
      typeof optionProps.children === "string"
        ? optionProps.children
        : String(optionProps.children ?? "");

    options.push({
      disabled: optionProps.disabled,
      label,
      value,
    });
  });

  return options;
}

function readMenuSide(node: HTMLElement | null): SelectMenuSide {
  const side = node?.getAttribute("data-side");
  return side === "top" ? "top" : "bottom";
}

function ChevronDownIcon({ className }: { className?: string }) {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={1.75}
      viewBox="0 0 24 24"
    >
      <path d="m6 9 6 6 6-6" />
    </svg>
  );
}

/** Custom select — Radix listbox; keeps native `<option>` children for RHF compatibility. */
export const Select = React.forwardRef<HTMLButtonElement, SelectProps>(
  (
    {
      "aria-invalid": ariaInvalid,
      children,
      className,
      disabled,
      id,
      name,
      onBlur,
      onChange,
      required,
      value,
    },
    ref,
  ) => {
    const contentRef = React.useRef<HTMLDivElement>(null);
    const [open, setOpen] = React.useState(false);
    const [menuSide, setMenuSide] = React.useState<SelectMenuSide>("bottom");
    const options = React.useMemo(() => parseSelectOptions(children), [children]);
    const placeholderOption = options.find((option) => option.value === "");
    const selectableOptions = options.filter((option) => option.value !== "");
    const stringValue =
      value === undefined || value === null ? "" : String(value);
    const resolvedValue = stringValue === "" ? undefined : stringValue;

    React.useLayoutEffect(() => {
      if (!open) {
        return;
      }

      const content = contentRef.current;
      if (!content) {
        return;
      }

      function syncMenuSide(): void {
        setMenuSide(readMenuSide(content));
      }

      syncMenuSide();

      const observer = new MutationObserver(syncMenuSide);
      observer.observe(content, {
        attributeFilter: ["data-side"],
        attributes: true,
      });

      return () => {
        observer.disconnect();
      };
    }, [open, selectableOptions.length]);

    function handleValueChange(nextValue: string): void {
      onChange?.({
        target: { name: name ?? "", value: nextValue },
        currentTarget: { name: name ?? "", value: nextValue },
      } as React.ChangeEvent<HTMLSelectElement>);
    }

    return (
      <SelectPrimitive.Root
        disabled={disabled}
        name={name}
        onOpenChange={setOpen}
        onValueChange={handleValueChange}
        open={open}
        required={required}
        value={resolvedValue}
      >
        <div className="field-select">
          <SelectPrimitive.Trigger
            ref={ref}
            aria-invalid={ariaInvalid}
            aria-required={required ? true : undefined}
            className={cn(
              "field-chrome field-chrome--select field-select-trigger",
              open && menuSide === "bottom" && "field-select-trigger--menu-bottom",
              open && menuSide === "top" && "field-select-trigger--menu-top",
              className,
            )}
            id={id}
            onBlur={(event) => {
              onBlur?.(event as unknown as React.FocusEvent<HTMLSelectElement>);
            }}
          >
            <SelectPrimitive.Value
              className="field-select-value"
              placeholder={placeholderOption?.label ?? "Select…"}
            />
            <SelectPrimitive.Icon asChild>
              <ChevronDownIcon className="field-select-chevron" />
            </SelectPrimitive.Icon>
          </SelectPrimitive.Trigger>

          <SelectPrimitive.Portal>
            <SelectPrimitive.Content
              ref={contentRef}
              align="start"
              avoidCollisions
              className="field-select-content"
              collisionPadding={8}
              position="popper"
              side="bottom"
              sideOffset={0}
            >
              <SelectPrimitive.Viewport className="field-select-viewport">
                <div className="field-select-tray">
                  {selectableOptions.map((option) => (
                    <SelectPrimitive.Item
                      key={option.value}
                      className="field-select-item"
                      disabled={option.disabled}
                      value={option.value}
                    >
                      <SelectPrimitive.ItemText>{option.label}</SelectPrimitive.ItemText>
                    </SelectPrimitive.Item>
                  ))}
                </div>
              </SelectPrimitive.Viewport>
            </SelectPrimitive.Content>
          </SelectPrimitive.Portal>
        </div>
      </SelectPrimitive.Root>
    );
  },
);
Select.displayName = "Select";
