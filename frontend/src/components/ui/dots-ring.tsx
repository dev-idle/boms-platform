import type { ComponentProps } from "react";

import { cn } from "@/lib/utils";

type DotsRingProps = ComponentProps<"span"> & {
  dots?: number;
  dotScale?: number;
  label?: string;
  radiusScale?: number;
};

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max);
}

/** @loading-ui/dots-ring — installed via shadcn registry. */
function DotsRing({
  className,
  style,
  dots = 8,
  dotScale = 0.2,
  label,
  radiusScale = 0.34,
  ...props
}: DotsRingProps) {
  const dotCount = Math.max(4, Math.floor(dots));
  const safeDotScale = clamp(dotScale, 0.2, 0.4);
  const safeRadiusScale = clamp(radiusScale, 0, 0.5 - safeDotScale / 2);

  return (
    <span
      className={cn(
        "@container-[size] relative inline-flex aspect-square items-center justify-center",
        className,
      )}
      style={style}
      {...props}
    >
      <span aria-hidden="true" className="relative block size-full">
        {Array.from({ length: dotCount }, (_, index) => {
          const angle = (index / dotCount) * Math.PI * 2;
          const x = `${(Math.sin(angle) * safeRadiusScale * 100).toFixed(2)}cqmin`;
          const y = `${(-Math.cos(angle) * safeRadiusScale * 100).toFixed(2)}cqmin`;

          return (
            <span
              key={index}
              className="absolute top-1/2 left-1/2"
              style={{
                width: `calc(${safeDotScale} * 100cqmin)`,
                height: `calc(${safeDotScale} * 100cqmin)`,
                transform: `translate(-50%, -50%) translate(${x}, ${y})`,
              }}
            >
              <span
                className="loading-ui-dots-ring__dot block size-full rounded-full bg-current"
                style={{
                  animationDelay: `calc(var(--duration, 1s) / ${dotCount} * ${index - dotCount})`,
                }}
              />
            </span>
          );
        })}
      </span>
      {label ? <span className="sr-only">{label}</span> : null}
    </span>
  );
}

export { DotsRing };
