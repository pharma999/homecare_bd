package models

import "time"

type AppointmentType string
type AppointmentStatus string

const (
	AppointmentTypeHomeVisit AppointmentType = "HOME_VISIT"
	AppointmentTypeOnline    AppointmentType = "ONLINE"
	AppointmentTypeQuick     AppointmentType = "QUICK"
	AppointmentTypeScheduled AppointmentType = "SCHEDULED"
	AppointmentTypeEmergency AppointmentType = "EMERGENCY"

	AppointmentStatusPending    AppointmentStatus = "PENDING"
	AppointmentStatusConfirmed  AppointmentStatus = "CONFIRMED"
	AppointmentStatusInProgress AppointmentStatus = "IN_PROGRESS"
	AppointmentStatusCompleted  AppointmentStatus = "COMPLETED"
	AppointmentStatusCancelled  AppointmentStatus = "CANCELLED"
	AppointmentStatusNoShow     AppointmentStatus = "NO_SHOW"
)

type Appointment struct {
	ID               string            `bson:"_id,omitempty"               json:"appointment_id"`
	PatientUserID    string            `bson:"patient_user_id"             json:"patient_user_id"`
	DoctorID         string            `bson:"doctor_id,omitempty"         json:"doctor_id,omitempty"`
	NurseID          string            `bson:"nurse_id,omitempty"          json:"nurse_id,omitempty"`
	ServiceID        string            `bson:"service_id"                  json:"service_id"`
	AppointmentType  AppointmentType   `bson:"appointment_type"            json:"appointment_type"`
	Status           AppointmentStatus `bson:"status"                      json:"status"`
	ScheduledAt      time.Time         `bson:"scheduled_at"                json:"scheduled_at"`
	Duration         int               `bson:"duration"                    json:"duration"` // minutes
	PatientAddress   string            `bson:"patient_address,omitempty"   json:"patient_address,omitempty"`
	PatientLatitude  string            `bson:"patient_latitude,omitempty"  json:"patient_latitude,omitempty"`
	PatientLongitude string            `bson:"patient_longitude,omitempty" json:"patient_longitude,omitempty"`
	Notes            string            `bson:"notes,omitempty"             json:"notes,omitempty"`
	ProviderNotes    string            `bson:"provider_notes,omitempty"    json:"provider_notes,omitempty"`
	TotalAmount      float64           `bson:"total_amount"                json:"total_amount"`
	PaymentStatus    string            `bson:"payment_status"              json:"payment_status"`
	MeetingLink      string            `bson:"meeting_link,omitempty"      json:"meeting_link,omitempty"`
	CancelledReason  string            `bson:"cancelled_reason,omitempty"  json:"cancelled_reason,omitempty"`
	CreatedAt        time.Time         `bson:"created_at"                  json:"created_at"`
	UpdatedAt        time.Time         `bson:"updated_at"                  json:"updated_at"`
}

type Prescription struct {
	ID            string     `bson:"_id,omitempty"              json:"prescription_id"`
	AppointmentID string     `bson:"appointment_id"             json:"appointment_id"`
	DoctorID      string     `bson:"doctor_id"                  json:"doctor_id"`
	PatientUserID string     `bson:"patient_user_id"            json:"patient_user_id"`
	Medications   string     `bson:"medications"                json:"medications"` // JSON string
	Instructions  string     `bson:"instructions,omitempty"     json:"instructions,omitempty"`
	FollowUpDate  *time.Time `bson:"follow_up_date,omitempty"   json:"follow_up_date,omitempty"`
	IssuedAt      time.Time  `bson:"issued_at"                  json:"issued_at"`
	CreatedAt     time.Time  `bson:"created_at"                 json:"created_at"`
}

// ---- DTOs ----

type CreateAppointmentRequest struct {
	DoctorID         string          `json:"doctor_id"`
	NurseID          string          `json:"nurse_id"`
	ServiceID        string          `json:"service_id"        binding:"required"`
	AppointmentType  AppointmentType `json:"appointment_type"  binding:"required"`
	ScheduledAt      time.Time       `json:"scheduled_at"      binding:"required"`
	PatientAddress   string          `json:"patient_address"`
	PatientLatitude  string          `json:"patient_latitude"`
	PatientLongitude string          `json:"patient_longitude"`
	Notes            string          `json:"notes"`
}

type UpdateAppointmentStatusRequest struct {
	Status          AppointmentStatus `json:"status"           binding:"required"`
	ProviderNotes   string            `json:"provider_notes"`
	CancelledReason string            `json:"cancelled_reason"`
}
