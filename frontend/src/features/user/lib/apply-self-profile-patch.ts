import { USER_ROLE } from "@/constants/roles";

import type { UpdateSelfProfileInput } from "../schemas/index";
import type { Me } from "../types";

/** Applies PATCH /me fields onto cached `Me` for optimistic updates. */
export function applySelfProfilePatch(
  current: Me,
  input: UpdateSelfProfileInput,
): Me {
  switch (current.role) {
    case USER_ROLE.customer:
      return {
        ...current,
        profile: {
          ...current.profile,
          display_name:
            input.display_name === undefined
              ? current.profile.display_name
              : input.display_name,
          phone:
            input.phone === undefined ? current.profile.phone : input.phone,
        },
      };
    case USER_ROLE.staff:
    case USER_ROLE.baker:
    case USER_ROLE.manager:
      return {
        ...current,
        profile: {
          ...current.profile,
          full_name:
            input.full_name === undefined
              ? current.profile.full_name
              : input.full_name,
          phone:
            input.phone === undefined ? current.profile.phone : input.phone,
        },
      };
    case USER_ROLE.admin:
      return {
        ...current,
        profile: {
          ...current.profile,
          full_name:
            input.full_name === undefined
              ? current.profile.full_name
              : input.full_name,
          phone:
            input.phone === undefined ? current.profile.phone : input.phone,
        },
      };
    default: {
      const _exhaustive: never = current;
      return _exhaustive;
    }
  }
}
