-- Enable auth-related PostgreSQL extensions (Neon supports both).
CREATE EXTENSION IF NOT EXISTS "citext" WITH SCHEMA "public" VERSION "1.6";
CREATE EXTENSION IF NOT EXISTS "pgcrypto" WITH SCHEMA "public" VERSION "1.3";
-- Create enum type "user_role"
CREATE TYPE "user_role" AS ENUM ('customer', 'admin');
