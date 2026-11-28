"use client";

type QuantityStepperProps = {
  /** What the quantity is of, for the buttons' names. */
  itemName: string;
  value: number;
  min: number;
  max: number;
  onChange: (next: number) => void;
  /**
   * A change is under way: the steps ignore presses until it lands but stay
   * enabled, so the one just pressed keeps the keyboard focus.
   */
  pending: boolean;
  /** Why the line cannot change at all, when it cannot; the steps are disabled with it as their title. */
  lockedReason?: string;
};

/** One box: less, the quantity, more. */
export function QuantityStepper({ itemName, value, min, max, onChange, pending, lockedReason }: QuantityStepperProps) {
  function reason(atBound: string | undefined): string | undefined {
    return lockedReason ?? (pending ? "Updating the quantity…" : atBound);
  }

  function step(next: number): void {
    if (!pending) {
      onChange(next);
    }
  }

  return (
    <div aria-label={`Quantity of ${itemName}`} className="quantity-stepper" role="group">
      <button
        aria-disabled={pending || undefined}
        aria-label={`One less ${itemName}`}
        className="quantity-stepper__step"
        disabled={lockedReason !== undefined || value <= min}
        onClick={() => step(value - 1)}
        title={reason(value <= min ? "Remove the item to take it out of the cart" : undefined)}
        type="button"
      >
        −
      </button>
      <output aria-live="polite" className="quantity-stepper__value">
        {value}
      </output>
      <button
        aria-disabled={pending || undefined}
        aria-label={`One more ${itemName}`}
        className="quantity-stepper__step"
        disabled={lockedReason !== undefined || value >= max}
        onClick={() => step(value + 1)}
        title={reason(value >= max ? `At most ${max} of one item` : undefined)}
        type="button"
      >
        +
      </button>
    </div>
  );
}
