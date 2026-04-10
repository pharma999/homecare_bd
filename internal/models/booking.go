package models

import "time"

type BookingStatus string

const (
	BookingStatusPending    BookingStatus = "PENDING"
	BookingStatusAccepted   BookingStatus = "ACCEPTED"
	BookingStatusRejected   BookingStatus = "REJECTED"
	BookingStatusInProgress BookingStatus = "IN_PROGRESS"
	BookingStatusCompleted  BookingStatus = "COMPLETED"
	BookingStatusCancelled  BookingStatus = "CANCELLED"
)

type Booking struct {
	ID               string        `bson:"_id,omitempty"              json:"booking_id"`
	PatientUserID    string        `bson:"patient_user_id"            json:"patient_user_id"`
	ProfessionalID   string        `bson:"professional_id,omitempty"  json:"professional_id,omitempty"`
	ServiceID        string        `bson:"service_id"                 json:"service_id"`
	BookingType      string        `bson:"booking_type"               json:"booking_type"` // QUICK | SCHEDULED
	Status           BookingStatus `bson:"status"                     json:"status"`
	ScheduledAt      *time.Time    `bson:"scheduled_at,omitempty"     json:"scheduled_at,omitempty"`
	EstimatedArrival string        `bson:"estimated_arrival,omitempty" json:"estimated_arrival,omitempty"`
	PatientAddress   string        `bson:"patient_address"            json:"patient_address"`
	PatientLatitude  string        `bson:"patient_latitude,omitempty" json:"patient_latitude,omitempty"`
	PatientLongitude string        `bson:"patient_longitude,omitempty" json:"patient_longitude,omitempty"`
	TotalAmount      float64       `bson:"total_amount"               json:"total_amount"`
	PaymentStatus    string        `bson:"payment_status"             json:"payment_status"`
	Notes            string        `bson:"notes,omitempty"            json:"notes,omitempty"`
	CompletedAt      *time.Time    `bson:"completed_at,omitempty"     json:"completed_at,omitempty"`
	CancelledReason  string        `bson:"cancelled_reason,omitempty" json:"cancelled_reason,omitempty"`
	CreatedAt        time.Time     `bson:"created_at"                 json:"created_at"`
	UpdatedAt        time.Time     `bson:"updated_at"                 json:"updated_at"`
}

type CartItem struct {
	ID        string    `bson:"_id,omitempty" json:"cart_item_id"`
	UserID    string    `bson:"user_id"       json:"user_id"`
	ServiceID string    `bson:"service_id"    json:"service_id"`
	Title     string    `bson:"title"         json:"title"`
	Price     float64   `bson:"price"         json:"price"`
	Quantity  int       `bson:"quantity"      json:"quantity"`
	CreatedAt time.Time `bson:"created_at"    json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"    json:"updated_at"`
}

type Payment struct {
	ID            string     `bson:"_id,omitempty"              json:"payment_id"`
	UserID        string     `bson:"user_id"                    json:"user_id"`
	BookingID     string     `bson:"booking_id,omitempty"       json:"booking_id,omitempty"`
	AppointmentID string     `bson:"appointment_id,omitempty"   json:"appointment_id,omitempty"`
	Amount        float64    `bson:"amount"                     json:"amount"`
	Currency      string     `bson:"currency"                   json:"currency"`
	PaymentMethod string     `bson:"payment_method"             json:"payment_method"`
	TransactionID string     `bson:"transaction_id,omitempty"   json:"transaction_id,omitempty"`
	Status        string     `bson:"status"                     json:"status"` // PENDING|SUCCESS|FAILED|REFUNDED
	PaidAt        *time.Time `bson:"paid_at,omitempty"          json:"paid_at,omitempty"`
	Metadata      string     `bson:"metadata,omitempty"         json:"metadata,omitempty"`
	CreatedAt     time.Time  `bson:"created_at"                 json:"created_at"`
	UpdatedAt     time.Time  `bson:"updated_at"                 json:"updated_at"`
}

type Review struct {
	ID        string    `bson:"_id,omitempty"          json:"review_id"`
	UserID    string    `bson:"user_id"                json:"user_id"`
	BookingID string    `bson:"booking_id,omitempty"   json:"booking_id,omitempty"`
	DoctorID  string    `bson:"doctor_id,omitempty"    json:"doctor_id,omitempty"`
	NurseID   string    `bson:"nurse_id,omitempty"     json:"nurse_id,omitempty"`
	Rating    int       `bson:"rating"                 json:"rating"` // 1–5
	Comment   string    `bson:"comment,omitempty"      json:"comment,omitempty"`
	IsPublic  bool      `bson:"is_public"              json:"is_public"`
	CreatedAt time.Time `bson:"created_at"             json:"created_at"`
}

// ---- DTOs ----

type CreateBookingRequest struct {
	ProfessionalID   string  `json:"professional_id"`
	ServiceID        string  `json:"service_id"        binding:"required"`
	BookingType      string  `json:"booking_type"      binding:"required"`
	ScheduledAt      *string `json:"scheduled_at"`
	PatientAddress   string  `json:"patient_address"   binding:"required"`
	PatientLatitude  string  `json:"patient_latitude"`
	PatientLongitude string  `json:"patient_longitude"`
	Notes            string  `json:"notes"`
}

type AddToCartRequest struct {
	ServiceID string  `json:"service_id" binding:"required"`
	Title     string  `json:"title"      binding:"required"`
	Price     float64 `json:"price"      binding:"required"`
	Quantity  int     `json:"quantity"`
}

type UpdateCartQuantityRequest struct {
	ServiceID string `json:"service_id" binding:"required"`
	Quantity  int    `json:"quantity"   binding:"required"`
}
