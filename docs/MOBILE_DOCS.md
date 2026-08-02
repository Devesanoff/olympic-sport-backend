# Mobile Documentation (Scanner App)

This document provides all API details and offline architecture specifications for the Mobile Scanner App (React Native/Flutter).

## 1. Overview & Base Configuration
- **Base URL:** `http://localhost:8080/api`
- **Authentication:** Scanners authenticate via a Bearer JWT Token.
- **Default Operator Credentials:**
  - **Email:** `scanner@olympic.com` (Static testing account configured in backend)
  - **Password:** `scanner123`

## 2. Scanner Gate APIs (Online Mode)

When the mobile device has a stable internet connection, it must validate scans directly against the backend.

### `POST /api/scan/access`
Validates a participant's entry or exit at a specific zone.

**Request Body:**
```json
{
  "qr_token": "uuid.timestamp.signature",
  "zone_id": 1,
  "direction": "IN" // or "OUT"
}
```

**Response (200 OK):**
*Note: A 200 OK is returned even if access is denied. Check the `status` field.*
```json
{
  "status": "ALLOWED", // or "DENIED"
  "participant_id": "123e4567-e89b-12d3-a456-426614174000",
  "participant_name": "Usain Bolt",
  "zone_id": 1,
  "direction": "IN",
  "occupancy_count": 145,
  "reason": "" // Present if denied, e.g., "invalid QR signature"
}
```

### `POST /api/scan/meal`
Validates and claims a meal for a participant.

**Request Body:**
```json
{
  "qr_token": "uuid.timestamp.signature"
}
```

**Response (200 OK):**
```json
{
  "status": "ALLOWED", // or "DENIED"
  "participant_id": "123e4567-e89b-12d3-a456-426614174000",
  "participant_name": "Usain Bolt",
  "meal_type": "LUNCH",
  "reason": "" // Present if denied, e.g., "meal already claimed today"
}
```

## 3. Offline Architecture & Sync Spec

The scanner app must operate seamlessly during network outages.

### `GET /api/sync/offline-package`
The app must call this on initial login or manual sync to download the complete state of the accreditation system.

**Response Schema:**
```json
{
  "categories": [
    { "id": 1, "name": "Athlete", "color_code": "#00FF00", "can_eat": true, "created_at": "..." }
  ],
  "zones": [
    { "id": 1, "name": "Olympic Village", "code": "ZONE_A", "requires_in_out": true, "created_at": "..." }
  ],
  "category_allowed_zones": [
    { "category_id": 1, "zone_id": 1 }
  ],
  "meal_schedules": [
    { "id": 1, "date": "2026-08-02", "meal_type": "LUNCH", "start_time": "11:30:00", "end_time": "14:30:00" }
  ],
  "participants": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "category_id": 1,
      "status": "ACTIVE",
      "qr_token": "123e4567-e89b-12d3-a456-426614174000.1817209502.abcdef123456..."
    }
  ]
}
```

### `POST /api/sync/upload-logs`
When the app reconnects to the internet, it must bulk upload offline scans.

**Request Body:**
```json
{
  "access_logs": [
    {
      "id": "uuid",
      "participant_id": "participant-uuid",
      "zone_id": 1,
      "direction": "IN",
      "status": "ALLOWED",
      "reason": "",
      "created_at": "2026-08-02T12:00:00Z"
    }
  ],
  "meal_logs": [
    {
      "id": "uuid",
      "participant_id": "participant-uuid",
      "meal_schedule_id": 1,
      "meal_type": "LUNCH",
      "date": "2026-08-02",
      "status": "ALLOWED",
      "reason": "",
      "created_at": "2026-08-02T12:05:00Z"
    }
  ]
}
```

## 4. HMAC QR Verification Logic

When offline, the mobile app cannot ask the server to validate QR codes. It must validate the cryptographic signature itself.

**QR Token Structure:**
`PARTICIPANT_UUID.TIMESTAMP.SIGNATURE`

**Offline Validation Steps:**
1. Split the token by `.`. Ensure there are exactly 3 parts.
2. Form the validation string: `payload = PARTICIPANT_UUID + "." + TIMESTAMP`.
3. Compute the HMAC-SHA256 hash of the `payload` using the environment `HMAC_SECRET` key (compiled into the mobile app securely).
4. Hex-encode the computed HMAC.
5. Compare the computed Hex string with the `SIGNATURE` part from the QR token.
6. If they match exactly, the token is mathematically authentic and untampered.

## 5. Offline Storage Checklist
Use SQLite / WatermelonDB to store:
- [ ] `Categories` (For UI color rendering and meal capability checks).
- [ ] `Zones` (For dropdown selection in settings).
- [ ] `CategoryAllowedZones` (Join table for offline authorization logic).
- [ ] `MealSchedules` (To verify if a meal scan is occurring within an active window).
- [ ] `Participants` (To verify participant status, UUID, and look up their `category_id`).
- [ ] `PendingSyncLogs` (Table for caching successful and denied scans while offline).
