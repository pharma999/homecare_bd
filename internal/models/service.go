package models

import "time"

type ServiceCategory struct {
	ID           string    `bson:"_id,omitempty"           json:"category_id"`
	Name         string    `bson:"name"                    json:"name"`
	Slug         string    `bson:"slug"                    json:"slug"`
	Description  string    `bson:"description,omitempty"   json:"description,omitempty"`
	Icon         string    `bson:"icon,omitempty"          json:"icon,omitempty"`
	Color        string    `bson:"color,omitempty"         json:"color,omitempty"`
	IsActive     bool      `bson:"is_active"               json:"is_active"`
	DisplayOrder int       `bson:"display_order"           json:"display_order"`
	CreatedAt    time.Time `bson:"created_at"              json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"              json:"updated_at"`
}

type Service struct {
	ID           string    `bson:"_id,omitempty"           json:"service_id"`
	CategoryID   string    `bson:"category_id"             json:"category_id"`
	Title        string    `bson:"title"                   json:"title"`
	Slug         string    `bson:"slug"                    json:"slug"`
	Description  string    `bson:"description,omitempty"   json:"description,omitempty"`
	Icon         string    `bson:"icon,omitempty"          json:"icon,omitempty"`
	Color        string    `bson:"color,omitempty"         json:"color,omitempty"`
	BasePrice    float64   `bson:"base_price"              json:"base_price"`
	Duration     int       `bson:"duration"                json:"duration"` // minutes
	IsActive     bool      `bson:"is_active"               json:"is_active"`
	IsQuick      bool      `bson:"is_quick"                json:"is_quick"`
	IsEmergency  bool      `bson:"is_emergency"            json:"is_emergency"`
	DisplayOrder int       `bson:"display_order"           json:"display_order"`
	CreatedAt    time.Time `bson:"created_at"              json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"              json:"updated_at"`
}

type Professional struct {
	ID                 string    `bson:"_id,omitempty"                  json:"id"`
	UserID             string    `bson:"user_id"                        json:"user_id"`
	Role               string    `bson:"role"                           json:"role"`
	ServiceName        string    `bson:"service_name"                   json:"service_name"`
	Rating             float64   `bson:"rating"                         json:"rating"`
	IsAvailable        bool      `bson:"is_available"                   json:"available"`
	YearsOfExperience  int       `bson:"years_of_experience"            json:"years_experience"`
	EstimatedDuration  int       `bson:"estimated_duration"             json:"estimated_duration"`
	AvailableTimeStart string    `bson:"available_time_start,omitempty" json:"available_time_start,omitempty"`
	AvailableTimeEnd   string    `bson:"available_time_end,omitempty"   json:"available_time_end,omitempty"`
	Latitude           string    `bson:"latitude,omitempty"             json:"latitude,omitempty"`
	Longitude          string    `bson:"longitude,omitempty"            json:"longitude,omitempty"`
	CreatedAt          time.Time `bson:"created_at"                     json:"created_at"`
	UpdatedAt          time.Time `bson:"updated_at"                     json:"updated_at"`
}

type ProfessionalResponse struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	Role               string  `json:"role"`
	ServiceName        string  `json:"service_name"`
	Distance           string  `json:"distance,omitempty"`
	Rating             float64 `json:"rating"`
	Available          bool    `json:"available"`
	ImageURL           string  `json:"image_url,omitempty"`
	YearsExperience    int     `json:"years_experience"`
	EstimatedDuration  int     `json:"estimated_duration"`
	AvailableTimeStart string  `json:"available_time_start,omitempty"`
	AvailableTimeEnd   string  `json:"available_time_end,omitempty"`
}
