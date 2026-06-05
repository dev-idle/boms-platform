-- Drop hire_date from staff_profiles (HR hire date tracking is out of scope).
ALTER TABLE "staff_profiles" DROP COLUMN IF EXISTS "hire_date";
