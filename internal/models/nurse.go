package models

import "time"

type NurseCategory string
type NurseApprovalStatus string

const (
	NurseCategoryGeneral        NurseCategory = "GENERAL"
	NurseCategoryICU            NurseCategory = "ICU"
	NurseCategoryPediatric      NurseCategory = "PEDIATRIC"
	NurseCategoryElderly        NurseCategory = "ELDERLY"
	NurseCategoryPostSurgery    NurseCategory = "POST_SURGERY"
	NurseCategoryMaternity      NurseCategory = "MATERNITY"
	NurseCategoryRehabilitation NurseCategory = "REHABILITATION"
	NurseCategoryEmergency      NurseCategory = "EMERGENCY"

	NurseApprovalPending  NurseApprovalStatus = "PENDING"
	NurseApprovalApproved NurseApprovalStatus = "APPROVED"
	NurseApprovalRejected NurseApprovalStatus = "REJECTED"
)

type Nurse struct {
	ID                string              `bson:"_id,omitempty"              json:"nurse_id"`
	UserID            string              `bson:"user_id"                    json:"user_id"`
	HospitalID        string              `bson:"hospital_id,omitempty"      json:"hospital_id,omitempty"`
	Qualification     string              `bson:"qualification,omitempty"    json:"qualification,omitempty"`
	Category          NurseCategory       `bson:"category"                   json:"category"`
	IDProofType       string              `bson:"id_proof_type"              json:"id_proof_type"`
	IDProofNumber     string              `bson:"id_proof_number"            json:"id_proof_number"`
	YearsOfExperience int                 `bson:"years_of_experience"        json:"years_of_experience"`
	ServiceArea       string              `bson:"service_area,omitempty"     json:"service_area,omitempty"`
	ShiftAvailability string              `bson:"shift_availability"         json:"shift_availability"` // "DAY"|"NIGHT"|"BOTH"
	EmergencyCapable  bool                `bson:"emergency_capable"          json:"emergency_capable"`
	HourlyRate        float64             `bson:"hourly_rate"                json:"hourly_rate"`
	DailyRate         float64             `bson:"daily_rate"                 json:"daily_rate"`
	IsAvailable       bool                `bson:"is_available"               json:"is_available"`
	ApprovalStatus    NurseApprovalStatus `bson:"approval_status"            json:"approval_status"`
	Rating            float64             `bson:"rating"                     json:"rating"`
	TotalReviews      int                 `bson:"total_reviews"              json:"total_reviews"`
	CreatedAt         time.Time           `bson:"created_at"                 json:"created_at"`
	UpdatedAt         time.Time           `bson:"updated_at"                 json:"updated_at"`
}

// ---- DTOs ----

type RegisterNurseRequest struct {
	UserID            string        `json:"user_id"            binding:"required"`
	HospitalID        string        `json:"hospital_id"`
	Qualification     string        `json:"qualification"`
	Category          NurseCategory `json:"category"`
	IDProofType       string        `json:"id_proof_type"      binding:"required"`
	IDProofNumber     string        `json:"id_proof_number"    binding:"required"`
	YearsOfExperience int           `json:"years_of_experience"`
	ServiceArea       string        `json:"service_area"`
	ShiftAvailability string        `json:"shift_availability"`
	EmergencyCapable  bool          `json:"emergency_capable"`
	HourlyRate        float64       `json:"hourly_rate"`
	DailyRate         float64       `json:"daily_rate"`
}

type NurseResponse struct {
	NurseID           string        `json:"nurse_id"`
	UserID            string        `json:"user_id"`
	Name              string        `json:"name"`
	ProfileImage      string        `json:"profile_image,omitempty"`
	Qualification     string        `json:"qualification,omitempty"`
	Category          NurseCategory `json:"category"`
	YearsOfExperience int           `json:"years_of_experience"`
	HourlyRate        float64       `json:"hourly_rate"`
	DailyRate         float64       `json:"daily_rate"`
	ShiftAvailability string        `json:"shift_availability"`
	EmergencyCapable  bool          `json:"emergency_capable"`
	IsAvailable       bool          `json:"is_available"`
	Rating            float64       `json:"rating"`
	TotalReviews      int           `json:"total_reviews"`
	HospitalName      string        `json:"hospital_name,omitempty"`
}
