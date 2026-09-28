import type { ReactNode } from "react";

import { BakerShell } from "@/components/layouts/baker-shell";
import { ThemeScope } from "@/components/theme/theme-scope";
import { APP_THEME } from "@/constants/themes";
import { BakerGate } from "@/features/auth";
import { BakerLiveUpdates } from "@/features/baker";

export default function BakerLayout({ children }: { children: ReactNode }) {
  return (
    <BakerGate>
      <BakerLiveUpdates />
      <ThemeScope theme={APP_THEME.dashboard}>
        <BakerShell>{children}</BakerShell>
      </ThemeScope>
    </BakerGate>
  );
}
