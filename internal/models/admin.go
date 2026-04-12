package models

import "time"

// ── Subscription Plan ──────────────────────────────────────────────────────

type PlanDuration string

const (
	PlanMonthly   PlanDuration = "MONTHLY"
	PlanQuarterly PlanDuration = "QUARTERLY"
	PlanYearly    PlanDuration = "YEARLY"
)

type SubscriptionPlan struct {
	ID           string       `bson:"_id,omitempty"   json:"id"`
	Name         string       `bson:"name"            json:"name"`
	Description  string       `bson:"description"     json:"description"`
	Duration     PlanDuration `bson:"duration"        json:"duration"`
	Price        float64      `bson:"price"           json:"price"`
	Features     []string     `bson:"features"        json:"features"`
	MaxBookings  int          `bson:"max_bookings"    json:"max_bookings"`
	MaxFamilyMembers int      `bson:"max_family_members" json:"max_family_members"`
	IsActive     bool         `bson:"is_active"       json:"is_active"`
	DisplayOrder int          `bson:"display_order"   json:"display_order"`
	CreatedAt    time.Time    `bson:"created_at"      json:"created_at"`
	UpdatedAt    time.Time    `bson:"updated_at"      json:"updated_at"`
}

// ── User Subscription ─────────────────────────────────────────────────────

type SubscriptionStatus string

const (
	SubStatusActive    SubscriptionStatus = "ACTIVE"
	SubStatusExpired   SubscriptionStatus = "EXPIRED"
	SubStatusCancelled SubscriptionStatus = "CANCELLED"
)

type UserSubscription struct {
	ID         string             `bson:"_id,omitempty" json:"id"`
	UserID     string             `bson:"user_id"       json:"user_id"`
	PlanID     string             `bson:"plan_id"       json:"plan_id"`
	PlanName   string             `bson:"plan_name"     json:"plan_name"`
	Status     SubscriptionStatus `bson:"status"        json:"status"`
	StartDate  time.Time          `bson:"start_date"    json:"start_date"`
	EndDate    time.Time          `bson:"end_date"      json:"end_date"`
	PaidAmount float64            `bson:"paid_amount"   json:"paid_amount"`
	CreatedAt  time.Time          `bson:"created_at"    json:"created_at"`
}

// ── Support / Complaint ───────────────────────────────────────────────────

type TicketStatus string
type TicketPriority string

const (
	TicketOpen       TicketStatus = "OPEN"
	TicketInProgress TicketStatus = "IN_PROGRESS"
	TicketResolved   TicketStatus = "RESOLVED"
	TicketClosed     TicketStatus = "CLOSED"

	TicketLow      TicketPriority = "LOW"
	TicketMedium   TicketPriority = "MEDIUM"
	TicketHigh     TicketPriority = "HIGH"
	TicketCritical TicketPriority = "CRITICAL"
)

type SupportTicket struct {
	ID           string         `bson:"_id,omitempty"      json:"id"`
	UserID       string         `bson:"user_id"            json:"user_id"`
	UserName     string         `bson:"user_name"          json:"user_name"`
	UserPhone    string         `bson:"user_phone"         json:"user_phone"`
	Subject      string         `bson:"subject"            json:"subject"`
	Description  string         `bson:"description"        json:"description"`
	Category     string         `bson:"category"           json:"category"` // BOOKING / PAYMENT / DOCTOR / GENERAL / EMERGENCY
	Status       TicketStatus   `bson:"status"             json:"status"`
	Priority     TicketPriority `bson:"priority"           json:"priority"`
	AssignedTo   string         `bson:"assigned_to"        json:"assigned_to"`
	Resolution   string         `bson:"resolution"         json:"resolution"`
	ReferenceID  string         `bson:"reference_id"       json:"reference_id"` // booking/appointment ID
	ResolvedAt   *time.Time     `bson:"resolved_at"        json:"resolved_at,omitempty"`
	CreatedAt    time.Time      `bson:"created_at"         json:"created_at"`
	UpdatedAt    time.Time      `bson:"updated_at"         json:"updated_at"`
}

// ── Platform Settings (Super Admin) ──────────────────────────────────────

type PlatformSettings struct {
	ID                    string    `bson:"_id,omitempty"             json:"id"`
	DoctorCommissionPct   float64   `bson:"doctor_commission_pct"     json:"doctor_commission_pct"`
	NurseCommissionPct    float64   `bson:"nurse_commission_pct"      json:"nurse_commission_pct"`
	BookingCommissionPct  float64   `bson:"booking_commission_pct"    json:"booking_commission_pct"`
	EmergencyBaseFee      float64   `bson:"emergency_base_fee"        json:"emergency_base_fee"`
	SupportEmail          string    `bson:"support_email"             json:"support_email"`
	SupportPhone          string    `bson:"support_phone"             json:"support_phone"`
	AppVersion            string    `bson:"app_version"               json:"app_version"`
	MaintenanceMode       bool      `bson:"maintenance_mode"          json:"maintenance_mode"`
	UpdatedBy             string    `bson:"updated_by"                json:"updated_by"`
	UpdatedAt             time.Time `bson:"updated_at"                json:"updated_at"`
}

// ── Service Zone ──────────────────────────────────────────────────────────

type ZoneStatus string

const (
	ZoneActive   ZoneStatus = "ACTIVE"
	ZoneInactive ZoneStatus = "INACTIVE"
	ZonePlanned  ZoneStatus = "PLANNED"
)

type ServiceZone struct {
	ID          string     `bson:"_id,omitempty" json:"id"`
	Name        string     `bson:"name"          json:"name"`
	City        string     `bson:"city"          json:"city"`
	State       string     `bson:"state"         json:"state"`
	PinCodes    []string   `bson:"pin_codes"     json:"pin_codes"`
	Status      ZoneStatus `bson:"status"        json:"status"`
	LaunchDate  *time.Time `bson:"launch_date"   json:"launch_date,omitempty"`
	CreatedBy   string     `bson:"created_by"    json:"created_by"`
	CreatedAt   time.Time  `bson:"created_at"    json:"created_at"`
	UpdatedAt   time.Time  `bson:"updated_at"    json:"updated_at"`
}

// ── Admin User (for super admin to manage admins) ─────────────────────────

type AdminCreateRequest struct {
	Name        string `json:"name"         binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Email       string `json:"email"`
	Role        string `json:"role" binding:"required"` // ADMIN or SUPER_ADMIN
}
