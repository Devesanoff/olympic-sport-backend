# Frontend Documentation (Nuxt 3 Admin Panel)

This documentation provides all necessary details to connect the frontend Nuxt 3 Admin Panel to the Olympic Sport Accreditation System Go Backend.

## 1. Overview & Base Configuration
- **Base URL:** `http://localhost:8080/api` (All endpoints are relative to this).
- **Authentication:** Most endpoints require a Bearer JWT Token.
- **Header Format:** `Authorization: Bearer <JWT_TOKEN>`
- **Content-Type:** `application/json`

## 2. Seed Credentials & Initial Testing
The database is seeded with a default SuperAdmin account. Use these credentials to obtain an administrator JWT:
- **Email:** `admin@olympic.org`
- **Password:** `AdminPassword123!`

## 3. Authentication APIs

### `POST /api/auth/login`
Authenticates a user and returns an access token.

**Request Body:**
```json
{
  "email": "admin@olympic.org",
  "password": "AdminPassword123!"
}
```

**Success Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI..."
}
```

## 4. Admin Management APIs (CRUD)

### Participants

**`GET /api/participants`**
List participants with pagination and optional filters.
- **Query Params:** `page` (default 1), `page_size` (default 20), `search` (name), `category_id`, `status`.

**`GET /api/participants/:id`**
Get a single participant by their UUID.

**`POST /api/participants`**
Create a new participant.
```json
{
  "full_name": "Usain Bolt",
  "category_id": 1,
  "status": "ACTIVE"
}
```

**`PUT /api/participants/:id`**
Update an existing participant.

**`DELETE /api/participants/:id`**
Delete a participant.

---
### Zones

**`GET /api/zones`**
List all venue zones.

**`POST /api/zones`**
```json
{
  "name": "Olympic Village Entrance",
  "code": "ZONE_A",
  "requires_in_out": true
}
```
*(Supports `GET /:id`, `PUT /:id`, `DELETE /:id`)*

---
### Categories

**`GET /api/categories`**
List all participant categories.

**`POST /api/categories`**
```json
{
  "name": "Athlete",
  "color_code": "#00FF00",
  "can_eat": true,
  "allowed_zone_ids": [1, 2, 3]
}
```
*(Supports `GET /:id`, `PUT /:id`, `DELETE /:id`)*

---
### Meal Schedules

**`GET /api/meal-schedules`**
List all meal schedules.

**`POST /api/meal-schedules`**
```json
{
  "date": "2026-08-02",
  "meal_type": "LUNCH",
  "start_time": "11:30:00",
  "end_time": "14:30:00",
  "allowed_category_ids": [1, 2]
}
```
*(Supports `GET /:id`, `PUT /:id`, `DELETE /:id`)*

---
### Users, Roles & Permissions

**`GET /api/users`** - List admins.
**`POST /api/users`** - Create admins, requires `email`, `password`, and `role_ids` array.
**`GET /api/roles`** - List system roles.
**`POST /api/roles`** - Create custom roles.
**`POST /api/roles/:id/permissions`** - Update dynamic RBAC permissions for a role.
**`GET /api/permissions`** - List all granular permissions available in the system.

## 5. Reports & Exports

### `GET /api/reports/access-logs`
Get paginated access scans.
- **Query Params:** `page`, `page_size`, `zone_id`, `direction`, `status`, `start_date`, `end_date`.

### `GET /api/reports/meal-logs`
Get paginated meal scans.
- **Query Params:** `page`, `page_size`, `meal_type`, `status`, `start_date`, `end_date`.

### `GET /api/reports/denied-attempts`
Gets aggregated repeated access denials to track potential unauthorized entries.

### `GET /api/reports/export/excel`
Generates and downloads a `.xlsx` report file containing filtered access or meal logs.
- **Query Params:** `report_type` (`access` or `meal`), plus filters.
- **Response:** Binary File (`application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`).

## 6. Badge Generation

### `GET /api/badges/:participantId/generate`
Generates a printable PDF Badge for a single participant containing their HMAC QR Code.
- **Response:** Binary File (`application/pdf`).

### `POST /api/badges/bulk-generate`
Generates a multi-page PDF badge file for a batch of participants.
```json
{
  "participant_ids": ["uuid-1", "uuid-2"]
}
```

## 7. Real-Time WebSocket Specification

The dashboard utilizes WebSockets to receive live zone occupancy and scan events.

- **Connection URL:** `ws://localhost:8080/api/dashboard/live-stats?token=<JWT_TOKEN>`

**Event Payload Schema (Emitted by Server):**
The server pushes JSON updates whenever a successful `IN` or `OUT` scan occurs.
```json
{
  "event_type": "zone_count_update",
  "timestamp": "2026-08-02T12:00:00Z",
  "data": {
    "zone_id": 1,
    "occupancy_count": 142,
    "total_entries": 450
  }
}
```

## 8. Standard Error Schema

The backend strictly adheres to a uniform JSON error format across all endpoints.

**Example 400 Bad Request / 422 Validation Error:**
```json
{
  "error": "Key: 'CreateParticipantRequest.FullName' Error:Field validation for 'FullName' failed on the 'required' tag"
}
```

**Example 401 Unauthorized (Invalid JWT):**
```json
{
  "error": "invalid or expired token"
}
```

**Example 403 Forbidden (RBAC Failure):**
```json
{
  "error": "insufficient permissions"
}
```

**Example 500 Internal Server Error:**
```json
{
  "error": "failed to generate authentication token"
}
```
