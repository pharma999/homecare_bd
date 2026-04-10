package models

import "time"

type DoctorApprovalStatus string
type ConsultationType string
type DoctorAvailability string

const (
	DoctorApprovalPending  DoctorApprovalStatus = "PENDING"
	DoctorApprovalApproved DoctorApprovalStatus = "APPROVED"
	DoctorApprovalRejected DoctorApprovalStatus = "REJECTED"

	ConsultationOnline    ConsultationType = "ONLINE"
	ConsultationHomeVisit ConsultationType = "HOME_VISIT"
	ConsultationBoth      ConsultationType = "BOTH"

	DoctorAvailOnline  DoctorAvailability = "ONLINE"
	DoctorAvailOffline DoctorAvailability = "OFFLINE"
	DoctorAvailBusy    DoctorAvailability = "BUSY"
)

type Doctor struct {
	ID                string               `bson:"_id,omitempty"              json:"doctor_id"`
	UserID            string               `bson:"user_id"                    json:"user_id"`
	HospitalID        string               `bson:"hospital_id,omitempty"      json:"hospital_id,omitempty"`
	Qualification     string               `bson:"qualification"              json:"qualification"`
	Specialty         string               `bson:"specialty"                  json:"specialty"`
	SubSpecialty      string               `bson:"sub_specialty,omitempty"    json:"sub_specialty,omitempty"`
	LicenseNumber     string               `bson:"license_number"             json:"license_number"`
	YearsOfExperience int                  `bson:"years_of_experience"        json:"years_of_experience"`
	Languages         string               `bson:"languages,omitempty"        json:"languages,omitempty"`
	ConsultationFee   float64              `bson:"consultation_fee"           json:"consultation_fee"`
	HomeVisitFee      float64              `bson:"home_visit_fee"             json:"home_visit_fee"`
	ConsultationType  ConsultationType     `bson:"consultation_type"          json:"consultation_type"`
	ServiceRadius     float64              `bson:"service_radius"             json:"service_radius"`
	Availability      DoctorAvailability   `bson:"availability"               json:"availability"`
	ApprovalStatus    DoctorApprovalStatus `bson:"approval_status"            json:"approval_status"`
	Bio               string               `bson:"bio,omitempty"              json:"bio,omitempty"`
	Rating            float64              `bson:"rating"                     json:"rating"`
	TotalReviews      int                  `bson:"total_reviews"              json:"total_reviews"`
	Schedules         []DoctorSchedule     `bson:"schedules,omitempty"        json:"schedules,omitempty"`
	CreatedAt         time.Time            `bson:"created_at"                 json:"created_at"`
	UpdatedAt         time.Time            `bson:"updated_at"                 json:"updated_at"`
}

type DoctorSchedule struct {
	DayOfWeek   int    `bson:"day_of_week"  json:"day_of_week"` // 0=Sun … 6=Sat
	StartTime   string `bson:"start_time"   json:"start_time"`  // "09:00"
	EndTime     string `bson:"end_time"     json:"end_time"`
	IsAvailable bool   `bson:"is_available" json:"is_available"`
}

// ---- DTOs ----

type RegisterDoctorRequest struct {
	UserID            string           `json:"user_id"            binding:"required"`
	HospitalID        string           `json:"hospital_id"`
	Qualification     string           `json:"qualification"      binding:"required"`
	Specialty         string           `json:"specialty"          binding:"required"`
	SubSpecialty      string           `json:"sub_specialty"`
	LicenseNumber     string           `json:"license_number"     binding:"required"`
	YearsOfExperience int              `json:"years_of_experience"`
	Languages         string           `json:"languages"`
	ConsultationFee   float64          `json:"consultation_fee"`
	HomeVisitFee      float64          `json:"home_visit_fee"`
	ConsultationType  ConsultationType `json:"consultation_type"`
	ServiceRadius     float64          `json:"service_radius"`
	Bio               string           `json:"bio"`
}

type UpdateDoctorRequest struct {
	Specialty         string             `json:"specialty"`
	SubSpecialty      string             `json:"sub_specialty"`
	ConsultationFee   float64            `json:"consultation_fee"`
	HomeVisitFee      float64            `json:"home_visit_fee"`
	ConsultationType  ConsultationType   `json:"consultation_type"`
	ServiceRadius     float64            `json:"service_radius"`
	Availability      DoctorAvailability `json:"availability"`
	Bio               string             `json:"bio"`
	Languages         string             `json:"languages"`
}

type DoctorSearchFilter struct {
	Specialty        string  `form:"specialty"`
	MaxDistance      float64 `form:"max_distance"`
	MinRating        float64 `form:"min_rating"`
	MaxFee           float64 `form:"max_fee"`
	Language         string  `form:"language"`
	ConsultationType string  `form:"consultation_type"`
	Page             int     `form:"page,default=1"`
	Limit            int     `form:"limit,default=20"`
}

type DoctorResponse struct {
	DoctorID          string             `json:"doctor_id"`
	UserID            string             `json:"user_id"`
	Name              string             `json:"name"`
	ProfileImage      string             `json:"profile_image,omitempty"`
	Qualification     string             `json:"qualification"`
	Specialty         string             `json:"specialty"`
	SubSpecialty      string             `json:"sub_specialty,omitempty"`
	YearsOfExperience int                `json:"years_of_experience"`
	ConsultationFee   float64            `json:"consultation_fee"`
	HomeVisitFee      float64            `json:"home_visit_fee"`
	ConsultationType  ConsultationType   `json:"consultation_type"`
	Availability      DoctorAvailability `json:"availability"`
	Rating            float64            `json:"rating"`
	TotalReviews      int                `json:"total_reviews"`
	Languages         string             `json:"languages,omitempty"`
	Bio               string             `json:"bio,omitempty"`
	HospitalName      string             `json:"hospital_name,omitempty"`
	ServiceRadius     float64            `json:"service_radius"`
}
