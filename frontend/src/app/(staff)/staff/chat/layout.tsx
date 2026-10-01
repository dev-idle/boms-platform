import type { ReactNode } from "react";

import { StaffChat } from "@/features/staff";

export default function StaffChatLayout({ children }: { children: ReactNode }) {
  return <StaffChat>{children}</StaffChat>;
}
