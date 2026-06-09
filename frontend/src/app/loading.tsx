import { ThemeScope } from "@/components/theme/theme-scope";
import { APP_THEME } from "@/constants/themes";

export default function Loading() {
  return (
    <ThemeScope theme={APP_THEME.storefront}>
      <div className="flex min-h-[40vh] items-center justify-center text-sm text-muted">
        Loading…
      </div>
    </ThemeScope>
  );
}
