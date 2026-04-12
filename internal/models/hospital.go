package models

import "time"

// GeoPoint is a GeoJSON Point for MongoDB 2dsphere geospatial queries.
// Coordinates are [longitude, latitude] per GeoJSON spec.
type GeoPoint struct {
	Type        string    `bson:"type"        json:"type"`        // always "Point"
	Coordinates []float64 `bson:"coordinates" json:"coordinates"` // [lng, lat]
}

// NewGeoPoint creates a GeoJSON Point from lat/lng values.
func NewGeoPoint(lat, lng float64) *GeoPoint {
	return &GeoPoint{Type: "Point", Coordinates: []float64{lng, lat}}
}

type HospitalApprovalStatus string

const (
	HospitalApprovalPending  HospitalApprovalStatus = "PENDING"
	HospitalApprovalApproved HospitalApprovalStatus = "APPROVED"
	HospitalApprovalRejected HospitalApprovalStatus = "REJECTED"
)

type Hospital struct {
	ID                 string                 `bson:"_id,omitempty"               json:"hospital_id"`
	UserID             string                 `bson:"user_id"                     json:"user_id"`
	Name               string                 `bson:"name"                        json:"name"`
	RegistrationNumber string                 `bson:"registration_number"         json:"registration_number"`
	Address            string                 `bson:"address"                     json:"address"`
	City               string                 `bson:"city"                        json:"city"`
	State              string                 `bson:"state,omitempty"             json:"state,omitempty"`
	PinCode            string                 `bson:"pin_code,omitempty"          json:"pin_code,omitempty"`
	Latitude           string                 `bson:"latitude,omitempty"          json:"latitude,omitempty"`
	Longitude          string                 `bson:"longitude,omitempty"         json:"longitude,omitempty"`
	Location           *GeoPoint              `bson:"location,omitempty"          json:"location,omitempty"`
	ServiceZone        string                 `bson:"service_zone,omitempty"      json:"service_zone,omitempty"`
	PhoneNumber        string                 `bson:"phone_number"                json:"phone_number"`
	Email              string                 `bson:"email,omitempty"             json:"email,omitempty"`
	Website            string                 `bson:"website,omitempty"           json:"website,omitempty"`
	HasEmergency       bool                   `bson:"has_emergency"               json:"has_emergency"`
	AmbulanceCount     int                    `bson:"ambulance_count"             json:"ambulance_count"`
	OperatingHours     string                 `bson:"operating_hours,omitempty"   json:"operating_hours,omitempty"`
	Specialties        string                 `bson:"specialties,omitempty"       json:"specialties,omitempty"`
	DistanceKm         float64                `bson:"distance_km,omitempty"       json:"distance_km,omitempty"`
	ApprovalStatus     HospitalApprovalStatus `bson:"approval_status"             json:"approval_status"`
	IsActive           bool                   `bson:"is_active"                   json:"is_active"`
	CreatedAt          time.Time              `bson:"created_at"                  json:"created_at"`
	UpdatedAt          time.Time              `bson:"updated_at"                  json:"updated_at"`
}

type Ambulance struct {
	ID               string    `bson:"_id,omitempty"               json:"ambulance_id"`
	HospitalID       string    `bson:"hospital_id"                 json:"hospital_id"`
	VehicleNumber    string    `bson:"vehicle_number"              json:"vehicle_number"`
	DriverName       string    `bson:"driver_name"                 json:"driver_name"`
	DriverPhone      string    `bson:"driver_phone"                json:"driver_phone"`
	AmbulanceType    string    `bson:"ambulance_type"              json:"ambulance_type"` // BASIC | ADVANCED | ICU
	IsAvailable      bool      `bson:"is_available"                json:"is_available"`
	CurrentLatitude  string    `bson:"current_latitude,omitempty"  json:"current_latitude,omitempty"`
	CurrentLongitude string    `bson:"current_longitude,omitempty" json:"current_longitude,omitempty"`
	Location         *GeoPoint `bson:"location,omitempty"          json:"location,omitempty"`
	EmergencyID      string    `bson:"emergency_id,omitempty"      json:"emergency_id,omitempty"`
	CreatedAt        time.Time `bson:"created_at"                  json:"created_at"`
	UpdatedAt        time.Time `bson:"updated_at"                  json:"updated_at"`
}

// ---- DTOs ----

type RegisterHospitalRequest struct {
	UserID             string `json:"user_id"              binding:"required"`
	Name               string `json:"name"                 binding:"required"`
	RegistrationNumber string `json:"registration_number"  binding:"required"`
	Address            string `json:"address"              binding:"required"`
	City               string `json:"city"                 binding:"required"`
	State              string `json:"state"`
	PinCode            string `json:"pin_code"`
	Latitude           string `json:"latitude"`
	Longitude          string `json:"longitude"`
	ServiceZone        string `json:"service_zone"`
	PhoneNumber        string `json:"phone_number"         binding:"required"`
	Email              string `json:"email"`
	HasEmergency       bool   `json:"has_emergency"`
	AmbulanceCount     int    `json:"ambulance_count"`
	OperatingHours     string `json:"operating_hours"`
	Specialties        string `json:"specialties"`
}
