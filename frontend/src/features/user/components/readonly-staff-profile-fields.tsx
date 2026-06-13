import { useId } from "react";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

import type { StaffProfile } from "../types";

type ReadonlyStaffProfileFieldsProps = {
  profile: StaffProfile;
};

export function ReadonlyStaffProfileFields({
  profile,
}: ReadonlyStaffProfileFieldsProps) {
  const employeeCodeId = useId();

  return (
    <div className="field-control field-control--readonly">
      <Label htmlFor={employeeCodeId}>Employee code</Label>
      <Input
        id={employeeCodeId}
        readOnly
        tabIndex={-1}
        value={profile.employee_code}
      />
    </div>
  );
}
