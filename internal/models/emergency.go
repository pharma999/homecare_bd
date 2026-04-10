package models

import "time"

type EmergencyStatus string
type EmergencyPriority string

const (
	EmergencyStatusTriggered  EmergencyStatus = "TRIGGERED"
	EmergencyStatusDispatched EmergencyStatus = "DISPATCHED"
	EmergencyStatusEnRoute    EmergencyStatus = "EN_ROUTE"
	EmergencyStatusArrived    EmergencyStatus = "ARRIVED"
	EmergencyStatusResolved   EmergencyStatus = "RESOLVED"
	EmergencyStatusCancelled  EmergencyStatus = "CANCELLED"

	EmergencyPriorityLow      EmergencyPriority = "LOW"
	EmergencyPriorityMedium   EmergencyPriority = "MEDIUM"
	EmergencyPriorityHigh     EmergencyPriority = "HIGH"
	EmergencyPriorityCritical EmergencyPriority = "CRITICAL"
)

type Emergency struct {
	ID                  string            `bson:"_id,omitempty"                    json:"emergency_id"`
	PatientUserID       string            `bson:"patient_user_id"                  json:"patient_user_id"`
	AmbulanceID         string            `bson:"ambulance_id,omitempty"           json:"ambulance_id,omitempty"`
	HospitalID          string            `bson:"hospital_id,omitempty"            json:"hospital_id,omitempty"`
	Status              EmergencyStatus   `bson:"status"                           json:"status"`
	Priority            EmergencyPriority `bson:"priority"                         json:"priority"`
	PatientLatitude     string            `bson:"patient_latitude"                 json:"patient_latitude"`
	PatientLongitude    string            `bson:"patient_longitude"                json:"patient_longitude"`
	PatientAddress      string            `bson:"patient_address,omitempty"        json:"patient_address,omitempty"`
	SymptomDescription  string            `bson:"symptom_description,omitempty"    json:"symptom_description,omitempty"`
	EmergencyType       string            `bson:"emergency_type,omitempty"         json:"emergency_type,omitempty"`
	DispatchedAt        *time.Time        `bson:"dispatched_at,omitempty"          json:"dispatched_at,omitempty"`
	ArrivedAt           *time.Time        `bson:"arrived_at,omitempty"             json:"arrived_at,omitempty"`
	ResolvedAt          *time.Time        `bson:"resolved_at,omitempty"            json:"resolved_at,omitempty"`
	HospitalNotified    bool              `bson:"hospital_notified"                json:"hospital_notified"`
	FamilyNotified      bool              `bson:"family_notified"                  json:"family_notified"`
	Notes               string            `bson:"notes,omitempty"                  json:"notes,omitempty"`
	CreatedAt           time.Time         `bson:"created_at"                       json:"created_at"`
	UpdatedAt           time.Time         `bson:"updated_at"                       json:"updated_at"`
}

type MedicalRecord struct {
	ID           string    `bson:"_id,omitempty"             json:"record_id"`
	UserID       string    `bson:"user_id"                   json:"user_id"`
	RecordType   string    `bson:"record_type"               json:"record_type"` // LAB_REPORT|PRESCRIPTION|XRAY|ECG|OTHER
	Title        string    `bson:"title"                     json:"title"`
	Description  string    `bson:"description,omitempty"     json:"description,omitempty"`
	FileURL      string    `bson:"file_url,omitempty"        json:"file_url,omitempty"`
	RecordDate   time.Time `bson:"record_date"               json:"record_date"`
	DoctorName   string    `bson:"doctor_name,omitempty"     json:"doctor_name,omitempty"`
	HospitalName string    `bson:"hospital_name,omitempty"   json:"hospital_name,omitempty"`
	IsShared     bool      `bson:"is_shared"                 json:"is_shared"`
	CreatedAt    time.Time `bson:"created_at"                json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"                json:"updated_at"`
}

type Notification struct {
	ID          string    `bson:"_id,omitempty"          json:"notification_id"`
	UserID      string    `bson:"user_id"                json:"user_id"`
	Title       string    `bson:"title"                  json:"title"`
	Body        string    `bson:"body"                   json:"body"`
	Type        string    `bson:"type"                   json:"type"` // BOOKING|APPOINTMENT|EMERGENCY|PAYMENT|GENERAL
	ReferenceID string    `bson:"reference_id,omitempty" json:"reference_id,omitempty"`
	IsRead      bool      `bson:"is_read"                json:"is_read"`
	CreatedAt   time.Time `bson:"created_at"             json:"created_at"`
}

// ---- DTOs ----

type SOSRequest struct {
	PatientLatitude    string `json:"patient_latitude"    binding:"required"`
	PatientLongitude   string `json:"patient_longitude"   binding:"required"`
	PatientAddress     string `json:"patient_address"`
	SymptomDescription string `json:"symptom_description"`
	EmergencyType      string `json:"emergency_type"`
}

type UpdateEmergencyStatusRequest struct {
	Status      EmergencyStatus   `json:"status"       binding:"required"`
	Priority    EmergencyPriority `json:"priority"`
	Notes       string            `json:"notes"`
	AmbulanceID string            `json:"ambulance_id"`
	HospitalID  string            `json:"hospital_id"`
}
