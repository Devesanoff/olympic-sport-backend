-- Drop Indexes
DROP INDEX IF EXISTS idx_meal_schedules_date_time;
DROP INDEX IF EXISTS idx_meal_logs_participant_date_meal;
DROP INDEX IF EXISTS idx_access_logs_created_at;
DROP INDEX IF EXISTS idx_access_logs_participant_direction;
DROP INDEX IF EXISTS idx_participants_qr_token;

-- Drop Tables (in reverse dependency order)
DROP TABLE IF EXISTS meal_logs;
DROP TABLE IF EXISTS access_logs;
DROP TABLE IF EXISTS meal_schedule_categories;
DROP TABLE IF EXISTS meal_schedules;
DROP TABLE IF EXISTS participants;
DROP TABLE IF EXISTS category_allowed_zones;
DROP TABLE IF EXISTS zones;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;

-- Drop Enums
DROP TYPE IF EXISTS meal_status_enum;
DROP TYPE IF EXISTS meal_type_enum;
DROP TYPE IF EXISTS access_status_enum;
DROP TYPE IF EXISTS direction_enum;
DROP TYPE IF EXISTS participant_status_enum;

-- Drop Extensions (optional/safe)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
-- DROP EXTENSION IF EXISTS "pgcrypto";
