import type { z } from "zod";

import {
  adminProfileSchema,
  customerProfileSchema,
  meSchema,
  staffProfileSchema,
} from "../schemas/index";

export type Me = z.infer<typeof meSchema>;
export type CustomerProfile = z.infer<typeof customerProfileSchema>;
export type StaffProfile = z.infer<typeof staffProfileSchema>;
export type AdminProfile = z.infer<typeof adminProfileSchema>;
