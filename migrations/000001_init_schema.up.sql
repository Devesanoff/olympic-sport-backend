-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enum Types
CREATE TYPE participant_status_enum AS ENUM ('ACTIVE', 'INACTIVE', 'SUSPENDED');
CREATE TYPE direction_enum AS ENUM ('IN', 'OUT');
CREATE TYPE access_status_enum AS ENUM ('ALLOWED', 'DENIED');
CREATE TYPE meal_type_enum AS ENUM ('BREAKFAST', 'LUNCH', 'DINNER', 'NIGHT_SNACK');
CREATE TYPE meal_status_enum AS ENUM ('ALLOWED', 'DENIED');

-- 1. RBAC Tables
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE role_permissions (
    role_id INT REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INT REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id INT REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- 2. Accreditation Tables
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    color_code VARCHAR(7) NOT NULL, -- Hex code, e.g. #FF5733
    can_eat BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE zones (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    code VARCHAR(20) UNIQUE NOT NULL, -- e.g. ZONE_A, ZONE_VIP
    requires_in_out BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE category_allowed_zones (
    category_id INT REFERENCES categories(id) ON DELETE CASCADE,
    zone_id INT REFERENCES zones(id) ON DELETE CASCADE,
    PRIMARY KEY (category_id, zone_id)
);

CREATE TABLE participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name VARCHAR(255) NOT NULL,
    category_id INT REFERENCES categories(id) ON DELETE RESTRICT,
    qr_token VARCHAR(512) UNIQUE NOT NULL, -- Encrypted or unique hash token
    status participant_status_enum NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Meal Management Tables
CREATE TABLE meal_schedules (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    meal_type meal_type_enum NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (date, meal_type)
);

CREATE TABLE meal_schedule_categories (
    meal_schedule_id INT REFERENCES meal_schedules(id) ON DELETE CASCADE,
    category_id INT REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (meal_schedule_id, category_id)
);

-- 4. Logs Tables
CREATE TABLE access_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    participant_id UUID REFERENCES participants(id) ON DELETE CASCADE,
    zone_id INT REFERENCES zones(id) ON DELETE RESTRICT,
    direction direction_enum NOT NULL,
    status access_status_enum NOT NULL,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE meal_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    participant_id UUID REFERENCES participants(id) ON DELETE CASCADE,
    meal_schedule_id INT REFERENCES meal_schedules(id) ON DELETE RESTRICT,
    meal_type meal_type_enum NOT NULL,
    date DATE NOT NULL,
    status meal_status_enum NOT NULL,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for High Performance & Low Latency Queries
CREATE INDEX idx_participants_qr_token ON participants(qr_token);
CREATE INDEX idx_access_logs_participant_direction ON access_logs(participant_id, direction, created_at DESC);
CREATE INDEX idx_access_logs_created_at ON access_logs(created_at DESC);
CREATE INDEX idx_meal_logs_participant_date_meal ON meal_logs(participant_id, date, meal_type);
CREATE INDEX idx_meal_schedules_date_time ON meal_schedules(date, start_time, end_time);
