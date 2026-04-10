package models

import "time"

type GenderType string
type UserStatus string
type BlockStatus string
type UserServiceStatus string
type ServiceStatusType string
type UserRole string

const (
	GenderMale   GenderType = "MALE"
	GenderFemale GenderType = "FEMALE"
	GenderOther  GenderType = "OTHER"

	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"

	BlockStatusUnblocked BlockStatus = "UNBLOCKED"
	BlockStatusBlocked   BlockStatus = "BLOCKED"

	UserServiceUnsubscribed UserServiceStatus = "UNSUBSCRIBED"
	UserServiceSubscribed   UserServiceStatus = "SUBSCRIBED"

	ServiceStatusNew    ServiceStatusType = "NEW"
	ServiceStatusActive ServiceStatusType = "ACTIVE"

	RolePatient    UserRole = "PATIENT"
	RoleFamily     UserRole = "FAMILY"
	RoleDoctor     UserRole = "DOCTOR"
	RoleNurse      UserRole = "NURSE"
	RoleCaregiver  UserRole = "CAREGIVER"
	RoleHospital   UserRole = "HOSPITAL"
	RoleAdmin      UserRole = "ADMIN"
	RoleSuperAdmin UserRole = "SUPER_ADMIN"
)

type User struct {
	ID            string            `bson:"_id,omitempty"         json:"user_id"`
	Name          string            `bson:"name"                  json:"name"`
	Email         string            `bson:"email,omitempty"       json:"email"`
	PhoneNumber   string            `bson:"phone_number"          json:"phone_number"`
	Gender        GenderType        `bson:"gender"                json:"gender"`
	DateOfBirth   *time.Time        `bson:"date_of_birth,omitempty" json:"date_of_birth,omitempty"`
	BloodGroup    string            `bson:"blood_group,omitempty" json:"blood_group,omitempty"`
	ProfileImage  string            `bson:"profile_image,omitempty" json:"profile_image,omitempty"`
	Role          UserRole          `bson:"role"                  json:"role"`
	Status        UserStatus        `bson:"status"                json:"status"`
	BlockStatus   BlockStatus       `bson:"block_status"          json:"block_status"`
	UserService   UserServiceStatus `bson:"user_service"          json:"user_service"`
	ServiceStatus ServiceStatusType `bson:"service_status"        json:"service_status"`
	Addresses     []Address         `bson:"addresses,omitempty"   json:"addresses,omitempty"`
	CreatedAt     time.Time         `bson:"created_at"            json:"created_at"`
	UpdatedAt     time.Time         `bson:"updated_at"            json:"updated_at"`
}

type Address struct {
	AddressID   string    `bson:"address_id"            json:"address_id"`
	HouseNumber string    `bson:"house_number"          json:"house_number"`
	Street      string    `bson:"street"                json:"street"`
	Landmark    string    `bson:"landmark,omitempty"    json:"landmark,omitempty"`
	City        string    `bson:"city,omitempty"        json:"city,omitempty"`
	State       string    `bson:"state,omitempty"       json:"state,omitempty"`
	PinCode     string    `bson:"pin_code"              json:"pin_code"`
	Latitude    string    `bson:"latitude,omitempty"    json:"latitude,omitempty"`
	Longitude   string    `bson:"longitude,omitempty"   json:"longitude,omitempty"`
	IsPrimary   bool      `bson:"is_primary"            json:"is_primary"`
	AddressType string    `bson:"address_type"          json:"address_type"` // "address1" | "address2"
	CreatedAt   time.Time `bson:"created_at"            json:"created_at"`
}

type OTP struct {
	ID          string    `bson:"_id,omitempty"`
	PhoneNumber string    `bson:"phone_number"`
	OTPCode     string    `bson:"otp_code"`
	IsUsed      bool      `bson:"is_used"`
	ExpiresAt   time.Time `bson:"expires_at"`
	CreatedAt   time.Time `bson:"created_at"`
}

type FamilyMember struct {
	ID            string    `bson:"_id,omitempty"          json:"family_member_id"`
	PatientUserID string    `bson:"patient_user_id"        json:"patient_user_id"`
	FamilyUserID  string    `bson:"family_user_id"         json:"family_user_id"`
	Relation      string    `bson:"relation"               json:"relation"`
	AccessLevel   string    `bson:"access_level"           json:"access_level"`
	AlertsEnabled bool      `bson:"alerts_enabled"         json:"alerts_enabled"`
	CreatedAt     time.Time `bson:"created_at"             json:"created_at"`
}

// ---- Request / Response DTOs ----

type UserResponse struct {
	UserID        string            `json:"user_id"`
	Name          string            `json:"name"`
	Email         string            `json:"email"`
	PhoneNumber   string            `json:"phone_number"`
	Gender        GenderType        `json:"gender"`
	Role          UserRole          `json:"role"`
	Status        UserStatus        `json:"status"`
	BlockStatus   BlockStatus       `json:"block_status"`
	UserService   UserServiceStatus `json:"user_service"`
	ServiceStatus ServiceStatusType `json:"service_status"`
	BloodGroup    string            `json:"blood_group,omitempty"`
	ProfileImage  string            `json:"profile_image,omitempty"`
	Address1      *AddressResponse  `json:"address_1,omitempty"`
	Address2      *AddressResponse  `json:"address_2,omitempty"`
}

type AddressResponse struct {
	HouseNumber string `json:"house_number"`
	Street      string `json:"street"`
	Landmark    string `json:"landmark,omitempty"`
	City        string `json:"city,omitempty"`
	PinCode     string `json:"pin_code"`
	Latitude    string `json:"latitude,omitempty"`
	Longitude   string `json:"longitude,omitempty"`
	IsPrimary   bool   `json:"is_primary"`
}

type UpdateProfileRequest struct {
	Name   string     `json:"name"`
	Email  string     `json:"email"`
	Gender GenderType `json:"gender"`
}

type UpdateAddressRequest struct {
	AddressType string `json:"addressType" binding:"required"`
	HouseNumber string `json:"houseNumber"`
	Street      string `json:"street"`
	Landmark    string `json:"landmark"`
	PinCode     string `json:"pinCode"`
	Latitude    string `json:"latitude"`
	Longitude   string `json:"longitude"`
}

type DeleteAccountRequest struct {
	UserID string `json:"userId" binding:"required"`
}
