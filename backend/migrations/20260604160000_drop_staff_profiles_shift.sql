-- Drop shift metadata from staff_profiles (shift scheduling is out of scope).
ALTER TABLE "staff_profiles" DROP COLUMN IF EXISTS "shift";
