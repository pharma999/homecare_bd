# Home Care Service — API Documentation

**Base URL:** `http://localhost:8080/api`  
**Content-Type:** `application/json`  
**Auth:** `Authorization: Bearer <token>` (JWT, HS256)

---

## Table of Contents

1. [Authentication](#1-authentication)
2. [User](#2-user)
3. [Family Members](#3-family-members)
4. [Doctors](#4-doctors)
5. [Nurses](#5-nurses)
6. [Hospitals & Ambulances](#6-hospitals--ambulances)
7. [Services & Categories](#7-services--categories)
8. [Professionals](#8-professionals)
9. [Cart](#9-cart)
10. [Bookings](#10-bookings)
11. [Payments](#11-payments)
12. [Appointments](#12-appointments)
13. [Prescriptions](#13-prescriptions)
14. [Emergency / SOS](#14-emergency--sos)
15. [Medical Records](#15-medical-records)
16. [Notifications](#16-notifications)
17. [Admin](#17-admin)
18. [Response Format](#18-response-format)
19. [Error Codes](#19-error-codes)

---

## 1. Authentication

All auth endpoints are **public** (no token required).

### POST `/api/login`

Send OTP to a phone number.

**Request**
```json
{
  "phone_number": "+919876543210"
}
```

**Response 200**
```json
{
  "status": "success",
  "message": "OTP sent successfully"
}
```

> In test mode (`OTP_TEST_MODE=true`), the OTP is always `5555`.

---

### POST `/api/verify`

Verify OTP and receive a JWT token. Creates a new user account if the phone number is new.

**Request**
```json
{
  "phone_number": "+919876543210",
  "otp": "5555"
}
```

**Response 200**
```json
{
  "status": "success",
  "message": "OTP verified successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "role": "PATIENT"
  }
}
```

> The `token` field is a **per-user JWT** generated at verify time, signed with `JWT_SECRET`. It is never stored server-side. Each login generates a fresh token.

---

## 2. User

All endpoints require `Authorization: Bearer <token>`.

### GET `/api/user/profile`

Get the authenticated user's profile.

**Response 200**
```json
{
  "status": "success",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Alex Morgan",
    "email": "alex@example.com",
    "phone_number": "+919876543210",
    "gender": "MALE",
    "role": "PATIENT",
    "status": "ACTIVE",
    "block_status": "UNBLOCKED",
    "user_service": "UNSUBSCRIBED",
    "service_status": "NEW",
    "address_1": {
      "house_number": "12A",
      "street": "Jankipuram Sector-H",
      "landmark": "Near City Mall",
      "pin_code": "226021",
      "latitude": "26.8467",
      "longitude": "80.9462",
      "is_primary": true
    },
    "address_2": null
  }
}
```

---

### PUT `/api/user/profile`

Update name, email, or gender.

**Request**
```json
{
  "name": "Alex Morgan",
  "email": "alex@example.com",
  "gender": "MALE"
}
```

**Response 200** — returns updated user profile (same shape as GET).

---

### PUT `/api/user/address`

Add or update an address slot.

**Request**
```json
{
  "address_type": "address1",
  "house_number": "12A",
  "street": "Jankipuram Sector-H",
  "landmark": "Near City Mall",
  "pin_code": "226021",
  "latitude": "26.8467",
  "longitude": "80.9462"
}
```

`address_type` is `"address1"` or `"address2"`.

---

### DELETE `/api/user/delete`

Permanently delete the authenticated user's account.

**Response 200**
```json
{ "status": "success", "message": "Account deleted successfully" }
```

---

## 3. Family Members

### GET `/api/user/family`

List family members linked to the account.

**Response 200**
```json
{
  "status": "success",
  "data": [
    {
      "id": "...",
      "name": "Jane Morgan",
      "relation": "SPOUSE",
      "age": 32,
      "blood_group": "B+",
      "medical_conditions": ["Diabetes"],
      "alerts_enabled": true
    }
  ]
}
```

---

### POST `/api/user/family`

Add a family member.

**Request**
```json
{
  "name": "Jane Morgan",
  "relation": "SPOUSE",
  "age": 32,
  "blood_group": "B+",
  "medical_conditions": ["Diabetes"],
  "alerts_enabled": true
}
```

---

## 4. Doctors

### GET `/api/doctors/search`

Search doctors with optional filters.

**Query Parameters**

| Param | Type | Description |
|---|---|---|
| `specialty` | string | e.g. `Cardiology` |
| `consultation_type` | string | `ONLINE` / `HOME_VISIT` / `BOTH` |
| `max_fee` | float | Maximum consultation fee |
| `language` | string | Preferred language |
| `rating` | float | Minimum rating |
| `page` | int | Page number (default 1) |
| `limit` | int | Results per page (default 10) |

**Response 200**
```json
{
  "status": "success",
  "data": [
    {
      "id": "...",
      "user_id": "...",
      "name": "Dr. Priya Sharma",
      "profile_image": "https://...",
      "specialty": "Cardiology",
      "qualification": "MBBS, MD",
      "experience": 8,
      "rating": 4.7,
      "total_reviews": 124,
      "consultation_fee": 500.0,
      "home_visit_fee": 800.0,
      "consultation_type": "BOTH",
      "availability": "ONLINE",
      "hospital_name": "Apollo Hospital",
      "languages": ["Hindi", "English"],
      "services": ["ECG", "Stress Test"]
    }
  ],
  "pagination": { "page": 1, "limit": 10, "total": 45 }
}
```

---

### GET `/api/doctors/:id`

Get a specific doctor's full profile including weekly schedule.

---

### GET `/api/doctors/:id/schedule`

Get doctor's weekly availability slots.

**Response 200**
```json
{
  "status": "success",
  "data": [
    { "day_of_week": "MONDAY", "start_time": "09:00", "end_time": "17:00", "is_available": true, "max_patients": 20 }
  ]
}
```

---

### POST `/api/doctors/register` 🔒

Register as a doctor (authenticated user).

**Request**
```json
{
  "specialty": "Cardiology",
  "qualification": "MBBS, MD",
  "license_number": "MCI-12345",
  "experience": 8,
  "consultation_fee": 500.0,
  "home_visit_fee": 800.0,
  "consultation_type": "BOTH",
  "languages": ["Hindi", "English"],
  "services": ["ECG", "Stress Test"]
}
```

---

## 5. Nurses

### GET `/api/nurses/search`

Search available nurses.

**Query Parameters:** `category`, `language`, `max_rate`, `page`, `limit`

**Nurse categories:** `GENERAL`, `ICU`, `PEDIATRIC`, `ELDER_CARE`, `POST_SURGICAL`, `HOME_CARE`

---

### GET `/api/nurses/:id`

Get nurse profile.

---

### POST `/api/nurses/register` 🔒

Register as a nurse.

**Request**
```json
{
  "category": "HOME_CARE",
  "qualification": "GNM",
  "license_number": "NCI-67890",
  "experience": 5,
  "hourly_rate": 300.0,
  "languages": ["Hindi"],
  "skills": ["Wound care", "IV therapy"]
}
```

---

## 6. Hospitals & Ambulances

### GET `/api/hospitals/nearby`

Find hospitals near a location.

**Query Parameters:** `latitude`, `longitude`, `radius` (km, default 10)

---

### GET `/api/hospitals/:id`

Get hospital details.

---

### GET `/api/hospitals/:id/doctors`

List doctors at a specific hospital.

---

### POST `/api/hospitals/register` 🔒

Register a hospital.

---

### POST `/api/hospitals/:id/ambulances` 🔒

Add an ambulance to a hospital fleet.

**Request**
```json
{
  "vehicle_number": "UP32-AB-1234",
  "vehicle_type": "ALS",
  "driver_name": "Ramesh Kumar",
  "driver_phone": "+919876543210"
}
```

---

### PUT `/api/ambulances/:id/location` 🔒

Update ambulance GPS coordinates (for tracking).

**Request**
```json
{ "latitude": 26.8467, "longitude": 80.9462 }
```

---

## 7. Services & Categories

### GET `/api/services/categories`

List all active service categories.

**Response 200**
```json
{
  "status": "success",
  "data": [
    { "id": "...", "name": "Nursing Care", "description": "...", "icon": "nursing", "is_active": true },
    { "id": "...", "name": "Physiotherapy", "description": "...", "icon": "physio", "is_active": true }
  ]
}
```

---

### GET `/api/services`

List all services.

---

### GET `/api/services/categories/:id/services`

List services within a category.

---

### GET `/api/services/:id`

Get service details.

**Response 200**
```json
{
  "status": "success",
  "data": {
    "id": "...",
    "category_id": "...",
    "name": "Home Nursing",
    "description": "Professional nursing care at home",
    "base_price": 600.0,
    "unit": "per visit",
    "image_url": "https://...",
    "rating": 4.5,
    "total_reviews": 89
  }
}
```

---

## 8. Professionals

### GET `/api/services/:id/professionals`

List professionals offering a specific service.

---

### GET `/api/professionals/:id`

Get professional profile.

---

### POST `/api/professionals/:id/review` 🔒

Submit a review for a professional.

**Request**
```json
{ "rating": 4.5, "comment": "Excellent service!" }
```

---

## 9. Cart

All cart endpoints require authentication.

### GET `/api/cart` 🔒

Get current cart contents.

**Response 200**
```json
{
  "status": "success",
  "data": [
    {
      "id": "...",
      "service_id": "...",
      "service_name": "Home Nursing",
      "price": 600.0,
      "quantity": 1,
      "professional_id": "...",
      "professional_name": "Nurse Rita",
      "scheduled_at": "2026-04-15T10:00:00Z"
    }
  ]
}
```

---

### POST `/api/cart` 🔒

Add item to cart.

**Request**
```json
{
  "service_id": "...",
  "quantity": 1,
  "professional_id": "...",
  "scheduled_at": "2026-04-15T10:00:00Z"
}
```

---

### PUT `/api/cart/:id` 🔒

Update cart item quantity.

**Request**
```json
{ "quantity": 2 }
```

---

### DELETE `/api/cart/:id` 🔒

Remove item from cart.

---

### POST `/api/cart/clear` 🔒

Remove all items from cart.

---

### POST `/api/cart/checkout` 🔒

Convert cart items into a booking.

**Request**
```json
{
  "address": "12A Jankipuram, Lucknow",
  "scheduled_at": "2026-04-15T10:00:00Z",
  "notes": "Please bring all equipment"
}
```

**Response 201**
```json
{
  "status": "success",
  "message": "Booking created from cart",
  "data": { "booking_id": "...", "total_amount": 600.0 }
}
```

---

## 10. Bookings

### POST `/api/bookings` 🔒

Create a direct booking (without cart).

**Request**
```json
{
  "service_id": "...",
  "professional_id": "...",
  "scheduled_at": "2026-04-15T10:00:00Z",
  "address": "12A Jankipuram, Lucknow",
  "notes": "Morning preferred"
}
```

**Response 201**
```json
{
  "status": "success",
  "data": {
    "id": "...",
    "service_name": "Home Nursing",
    "status": "PENDING",
    "total_amount": 600.0,
    "scheduled_at": "2026-04-15T10:00:00Z"
  }
}
```

**Booking statuses:** `PENDING` → `ACCEPTED` → `IN_PROGRESS` → `COMPLETED` | `REJECTED` | `CANCELLED`

---

### GET `/api/bookings` 🔒

Get user's bookings.

**Query Parameters:** `status`, `page`, `limit`

---

### GET `/api/bookings/:id` 🔒

Get booking details.

---

### POST `/api/bookings/:id/cancel` 🔒

Cancel a booking.

---

## 11. Payments

### POST `/api/payments/initiate` 🔒

Initiate a payment for a booking.

**Request**
```json
{
  "booking_id": "...",
  "amount": 600.0,
  "payment_method": "CARD"
}
```

**Response 200**
```json
{
  "status": "success",
  "data": {
    "payment_id": "...",
    "gateway_order_id": "...",
    "amount": 600.0
  }
}
```

---

### POST `/api/payments/confirm` 🔒

Confirm a payment after gateway callback.

**Request**
```json
{
  "payment_id": "...",
  "transaction_id": "TXN_1234567890"
}
```

---

### GET `/api/payments` 🔒

Get payment history.

**Response 200**
```json
{
  "status": "success",
  "data": [
    {
      "id": "...",
      "booking_id": "...",
      "amount": 600.0,
      "status": "COMPLETED",
      "payment_method": "CARD",
      "transaction_id": "TXN_1234567890",
      "created_at": "2026-04-11T10:00:00Z"
    }
  ]
}
```

---

## 12. Appointments

### POST `/api/appointments` 🔒

Book a doctor appointment.

**Request**
```json
{
  "doctor_id": "...",
  "type": "HOME_VISIT",
  "scheduled_at": "2026-04-15T09:00:00Z",
  "family_member_id": null,
  "address": "12A Jankipuram, Lucknow",
  "notes": "Blood pressure issue"
}
```

**Types:** `HOME_VISIT`, `ONLINE`, `QUICK`, `SCHEDULED`, `EMERGENCY`

**Response 201**
```json
{
  "status": "success",
  "data": {
    "id": "...",
    "doctor_name": "Dr. Priya Sharma",
    "doctor_specialty": "Cardiology",
    "type": "HOME_VISIT",
    "status": "PENDING",
    "scheduled_at": "2026-04-15T09:00:00Z",
    "fee": 800.0
  }
}
```

---

### GET `/api/appointments` 🔒

List user's appointments.

**Query Parameters:** `status`, `page`, `limit`

**Statuses:** `PENDING`, `CONFIRMED`, `IN_PROGRESS`, `COMPLETED`, `CANCELLED`, `NO_SHOW`

---

### GET `/api/appointments/:id` 🔒

Get appointment details.

---

### PUT `/api/appointments/:id/status` 🔒

Update appointment status (doctor/admin only).

**Request**
```json
{ "status": "CONFIRMED", "meeting_link": "https://meet.jit.si/xyz" }
```

---

### POST `/api/appointments/:id/cancel` 🔒

Cancel an appointment.

---

### GET `/api/appointments/family` 🔒

List appointments for all family members.

---

## 13. Prescriptions

### GET `/api/prescriptions` 🔒

Get all prescriptions for the authenticated patient.

**Response 200**
```json
{
  "status": "success",
  "data": [
    {
      "id": "...",
      "appointment_id": "...",
      "doctor_name": "Dr. Priya Sharma",
      "diagnosis": "Hypertension",
      "medicines": [
        {
          "name": "Amlodipine",
          "dosage": "5mg",
          "frequency": "Once daily",
          "duration": "30 days",
          "instructions": "Take after meals"
        }
      ],
      "notes": "Reduce sodium intake",
      "follow_up_date": "2026-05-15",
      "created_at": "2026-04-11T10:00:00Z"
    }
  ]
}
```

---

## 14. Emergency / SOS

### POST `/api/emergency/sos` 🔒

Trigger an SOS emergency. Notifies all family members with `alerts_enabled=true`.

**Request**
```json
{
  "description": "Chest pain, difficulty breathing",
  "address": "12A Jankipuram, Lucknow",
  "latitude": 26.8467,
  "longitude": 80.9462,
  "priority": "CRITICAL",
  "family_member_id": null
}
```

**Priorities:** `LOW`, `MEDIUM`, `HIGH`, `CRITICAL`

**Response 201**
```json
{
  "status": "success",
  "message": "Emergency SOS triggered",
  "data": {
    "id": "...",
    "status": "TRIGGERED",
    "priority": "CRITICAL",
    "created_at": "2026-04-11T10:05:00Z"
  }
}
```

---

### GET `/api/emergency` 🔒

List user's emergency history.

---

### GET `/api/emergency/active` 🔒

List all active emergencies (admin/hospital only).

---

### GET `/api/emergency/:id` 🔒

Get emergency details.

---

### PUT `/api/emergency/:id/status` 🔒

Update emergency status (hospital/admin).

**Request**
```json
{ "status": "DISPATCHED", "ambulance_id": "..." }
```

**Status flow:** `TRIGGERED` → `DISPATCHED` → `EN_ROUTE` → `ARRIVED` → `RESOLVED`

---

## 15. Medical Records

### GET `/api/medical-records` 🔒

Get authenticated user's medical records.

**Response 200**
```json
{
  "status": "success",
  "data": [
    {
      "id": "...",
      "title": "Blood Test Report",
      "type": "LAB_REPORT",
      "file_url": "https://...",
      "doctor_name": "Dr. Priya Sharma",
      "hospital_name": "Apollo Hospital",
      "record_date": "2026-03-15",
      "created_at": "2026-03-15T12:00:00Z"
    }
  ]
}
```

**Types:** `LAB_REPORT`, `PRESCRIPTION`, `SCAN`, `VACCINATION`, `OTHER`

---

### POST `/api/medical-records` 🔒

Upload a medical record (multipart/form-data or URL-based).

**Request**
```json
{
  "title": "Blood Test Report",
  "type": "LAB_REPORT",
  "file_url": "https://storage.example.com/report.pdf",
  "description": "Annual blood panel",
  "doctor_name": "Dr. Priya Sharma",
  "hospital_name": "Apollo Hospital",
  "record_date": "2026-03-15",
  "family_member_id": null
}
```

---

### DELETE `/api/medical-records/:id` 🔒

Delete a medical record.

---

### GET `/api/medical-records/family` 🔒

Get medical records for all family members.

---

## 16. Notifications

### GET `/api/user/notifications` 🔒

Get user notifications (newest first).

**Response 200**
```json
{
  "status": "success",
  "data": [
    {
      "id": "...",
      "title": "Appointment Confirmed",
      "message": "Your appointment with Dr. Priya Sharma is confirmed for 15 Apr.",
      "type": "APPOINTMENT",
      "is_read": false,
      "reference_id": "...",
      "created_at": "2026-04-11T10:00:00Z"
    }
  ]
}
```

---

### POST `/api/user/notifications/:id/read` 🔒

Mark a single notification as read.

---

### POST `/api/user/notifications/read-all` 🔒

Mark all notifications as read.

---

## 17. Admin

All admin endpoints require role `ADMIN` or `SUPER_ADMIN`.

### GET `/api/admin/analytics` 🔒🛡️

Platform analytics dashboard.

**Response 200**
```json
{
  "status": "success",
  "data": {
    "total_users": 1240,
    "total_doctors": 87,
    "total_nurses": 153,
    "total_hospitals": 24,
    "total_bookings": 3450,
    "total_appointments": 2100,
    "active_emergencies": 3,
    "total_revenue": 2345600.0
  }
}
```

---

### GET `/api/admin/users` 🔒🛡️

List all users with pagination.

---

### GET `/api/admin/users/:id` 🔒🛡️

Get any user's details.

---

### PUT `/api/admin/users/:id/block` 🔒🛡️

Block or unblock a user.

**Request**
```json
{ "block_status": "BLOCKED" }
```

---

### GET `/api/admin/doctors` 🔒🛡️

List all doctors with approval status filter.

**Query:** `?approval_status=PENDING`

---

### PUT `/api/admin/doctors/:id/approve` 🔒🛡️

Approve or reject a doctor registration.

**Request**
```json
{ "status": "APPROVED" }
```

---

### GET `/api/admin/nurses` 🔒🛡️

List all nurses.

---

### PUT `/api/admin/nurses/:id/approve` 🔒🛡️

Approve or reject a nurse registration.

---

### GET `/api/admin/hospitals` 🔒🛡️

List all hospitals.

---

### PUT `/api/admin/hospitals/:id/approve` 🔒🛡️

Approve or reject a hospital registration.

---

### POST `/api/admin/services/categories` 🔒🛡️

Create a service category.

**Request**
```json
{
  "name": "Physiotherapy",
  "description": "Physical therapy and rehabilitation services",
  "icon": "physio"
}
```

---

### POST `/api/admin/services` 🔒🛡️

Create a service.

**Request**
```json
{
  "category_id": "...",
  "name": "Home Physiotherapy",
  "description": "...",
  "base_price": 700.0,
  "unit": "per session"
}
```

---

### PUT `/api/admin/services/:id` 🔒🛡️

Update a service.

---

### GET `/api/admin/bookings` 🔒🛡️

List all bookings platform-wide.

---

### PUT `/api/admin/bookings/:id/status` 🔒🛡️

Update any booking status.

---

## 18. Response Format

All responses follow this envelope:

**Success**
```json
{
  "status": "success",
  "message": "Human-readable message",
  "data": { }
}
```

**Paginated Success**
```json
{
  "status": "success",
  "data": [ ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 45
  }
}
```

**Error**
```json
{
  "status": "error",
  "message": "Human-readable error",
  "error": "Technical detail (dev mode only)"
}
```

---

## 19. Error Codes

| HTTP | Meaning |
|---|---|
| 200 | OK |
| 201 | Created |
| 400 | Bad Request — missing or invalid fields |
| 401 | Unauthorized — missing or expired JWT |
| 403 | Forbidden — insufficient role |
| 404 | Not Found |
| 409 | Conflict — e.g. duplicate phone number |
| 500 | Internal Server Error |

---

## JWT Details

- **Algorithm:** HS256
- **Signing key:** `JWT_SECRET` environment variable (never stored per-user)
- **Token lifetime:** `JWT_EXPIRY_HOURS` (default 720 h = 30 days)
- **Claims:** `user_id`, `phone_number`, `role`, `exp`, `iat`
- **Per-user:** Every successful OTP verify generates a **unique token** for that user at that moment. No tokens are shared or stored server-side.

---

## Role Hierarchy

| Role | Description |
|---|---|
| `PATIENT` | Default role — can book services & appointments |
| `FAMILY` | Family member sub-account |
| `DOCTOR` | Can manage appointments, upload prescriptions |
| `NURSE` | Can manage bookings assigned to them |
| `CAREGIVER` | Similar to nurse, broader care scope |
| `HOSPITAL` | Can manage hospital, doctors, ambulances |
| `ADMIN` | Platform admin — approvals, analytics |
| `SUPER_ADMIN` | Full access |

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Server listen port |
| `GIN_MODE` | `debug` | `debug` or `release` |
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB connection string |
| `MONGO_DB_NAME` | `home_care_db` | Database name |
| `JWT_SECRET` | — | HMAC signing key (keep secret!) |
| `JWT_EXPIRY_HOURS` | `720` | Token lifetime in hours |
| `OTP_EXPIRY_MINUTES` | `10` | OTP validity window |
| `OTP_TEST_MODE` | `false` | Use fixed test OTP |
| `OTP_TEST_VALUE` | `5555` | Fixed OTP in test mode |
| `ALLOWED_ORIGINS` | `*` | CORS allowed origins |

---

*Generated for Home Care Service backend — Go 1.21 / Gin / MongoDB*
