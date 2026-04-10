package handlers

import (
	"context"
	"time"

	"home_care_backend/internal/database"
	"home_care_backend/internal/middleware"
	"home_care_backend/internal/models"
	"home_care_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterHospital(c *gin.Context) {
	var req models.RegisterHospitalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	ctx := context.Background()

	var user models.User
	if err := database.Col(database.ColUsers).FindOne(ctx, bson.M{"_id": req.UserID}).Decode(&user); err != nil {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	count, _ := database.Col(database.ColHospitals).CountDocuments(ctx, bson.M{"registration_number": req.RegistrationNumber})
	if count > 0 {
		utils.ErrorResponse(c, 409, "Registration number already exists", "duplicate_reg")
		return
	}

	now := time.Now()
	hospital := models.Hospital{
		ID:                 uuid.New().String(),
		UserID:             req.UserID,
		Name:               req.Name,
		RegistrationNumber: req.RegistrationNumber,
		Address:            req.Address,
		City:               req.City,
		State:              req.State,
		PinCode:            req.PinCode,
		Latitude:           req.Latitude,
		Longitude:          req.Longitude,
		ServiceZone:        req.ServiceZone,
		PhoneNumber:        req.PhoneNumber,
		Email:              req.Email,
		HasEmergency:       req.HasEmergency,
		AmbulanceCount:     req.AmbulanceCount,
		OperatingHours:     req.OperatingHours,
		Specialties:        req.Specialties,
		ApprovalStatus:     models.HospitalApprovalPending,
		IsActive:           true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if _, err := database.Col(database.ColHospitals).InsertOne(ctx, hospital); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to register hospital")
		return
	}
	database.Col(database.ColUsers).UpdateOne(ctx, bson.M{"_id": req.UserID},
		bson.M{"$set": bson.M{"role": models.RoleHospital}})
	utils.CreatedResponse(c, "Hospital registered. Pending admin approval.", hospital)
}

func GetHospital(c *gin.Context) {
	hospitalID := c.Param("hospitalId")
	var hospital models.Hospital
	err := database.Col(database.ColHospitals).FindOne(context.Background(), bson.M{"_id": hospitalID}).Decode(&hospital)
	if err == mongo.ErrNoDocuments {
		utils.NotFoundResponse(c, "Hospital not found")
		return
	}
	utils.SuccessResponse(c, "Hospital fetched", hospital)
}

func UpdateHospital(c *gin.Context) {
	hospitalID := c.Param("hospitalId")
	authUserID := middleware.GetUserID(c)
	ctx := context.Background()

	var hospital models.Hospital
	if err := database.Col(database.ColHospitals).FindOne(ctx, bson.M{"_id": hospitalID}).Decode(&hospital); err != nil {
		utils.NotFoundResponse(c, "Hospital not found")
		return
	}
	if hospital.UserID != authUserID {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	var req struct {
		Name           string `json:"name"`
		Address        string `json:"address"`
		PhoneNumber    string `json:"phone_number"`
		Email          string `json:"email"`
		HasEmergency   *bool  `json:"has_emergency"`
		AmbulanceCount *int   `json:"ambulance_count"`
		OperatingHours string `json:"operating_hours"`
		Specialties    string `json:"specialties"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	set := bson.M{"updated_at": time.Now()}
	if req.Name != "" {
		set["name"] = req.Name
	}
	if req.Address != "" {
		set["address"] = req.Address
	}
	if req.PhoneNumber != "" {
		set["phone_number"] = req.PhoneNumber
	}
	if req.Email != "" {
		set["email"] = req.Email
	}
	if req.HasEmergency != nil {
		set["has_emergency"] = *req.HasEmergency
	}
	if req.AmbulanceCount != nil {
		set["ambulance_count"] = *req.AmbulanceCount
	}
	if req.OperatingHours != "" {
		set["operating_hours"] = req.OperatingHours
	}
	if req.Specialties != "" {
		set["specialties"] = req.Specialties
	}

	database.Col(database.ColHospitals).UpdateOne(ctx, bson.M{"_id": hospitalID}, bson.M{"$set": set})
	database.Col(database.ColHospitals).FindOne(ctx, bson.M{"_id": hospitalID}).Decode(&hospital)
	utils.SuccessResponse(c, "Hospital updated", hospital)
}

func GetHospitalDoctors(c *gin.Context) {
	hospitalID := c.Param("hospitalId")
	ctx := context.Background()

	cursor, _ := database.Col(database.ColDoctors).Find(ctx,
		bson.M{"hospital_id": hospitalID, "approval_status": models.DoctorApprovalApproved})
	var doctors []models.Doctor
	cursor.All(ctx, &doctors)

	responses := make([]models.DoctorResponse, 0, len(doctors))
	for i := range doctors {
		responses = append(responses, buildDoctorResponse(ctx, &doctors[i]))
	}
	utils.SuccessResponse(c, "Hospital doctors fetched", responses)
}

func AddAmbulance(c *gin.Context) {
	hospitalID := c.Param("hospitalId")
	authUserID := middleware.GetUserID(c)
	ctx := context.Background()

	var hospital models.Hospital
	if err := database.Col(database.ColHospitals).FindOne(ctx, bson.M{"_id": hospitalID}).Decode(&hospital); err != nil {
		utils.NotFoundResponse(c, "Hospital not found")
		return
	}
	if hospital.UserID != authUserID {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	var req struct {
		VehicleNumber string `json:"vehicle_number" binding:"required"`
		DriverName    string `json:"driver_name"    binding:"required"`
		DriverPhone   string `json:"driver_phone"   binding:"required"`
		AmbulanceType string `json:"ambulance_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	now := time.Now()
	amb := models.Ambulance{
		ID:            uuid.New().String(),
		HospitalID:    hospitalID,
		VehicleNumber: req.VehicleNumber,
		DriverName:    req.DriverName,
		DriverPhone:   req.DriverPhone,
		AmbulanceType: req.AmbulanceType,
		IsAvailable:   true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	database.Col(database.ColAmbulances).InsertOne(ctx, amb)
	utils.CreatedResponse(c, "Ambulance added", amb)
}

func UpdateAmbulanceLocation(c *gin.Context) {
	ambulanceID := c.Param("ambulanceId")
	var req struct {
		Latitude    string `json:"latitude"     binding:"required"`
		Longitude   string `json:"longitude"    binding:"required"`
		IsAvailable *bool  `json:"is_available"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	set := bson.M{"current_latitude": req.Latitude, "current_longitude": req.Longitude, "updated_at": time.Now()}
	if req.IsAvailable != nil {
		set["is_available"] = *req.IsAvailable
	}
	database.Col(database.ColAmbulances).UpdateOne(context.Background(), bson.M{"_id": ambulanceID}, bson.M{"$set": set})
	utils.SuccessResponse(c, "Ambulance location updated", nil)
}

func GetNearbyHospitals(c *gin.Context) {
	city := c.Query("city")
	query := bson.M{"approval_status": models.HospitalApprovalApproved, "is_active": true}
	if city != "" {
		query["city"] = bson.M{"$regex": city, "$options": "i"}
	}

	cursor, _ := database.Col(database.ColHospitals).Find(context.Background(), query)
	var hospitals []models.Hospital
	cursor.All(context.Background(), &hospitals)
	utils.SuccessResponse(c, "Hospitals fetched", hospitals)
}

func GetHospitalEmergencies(c *gin.Context) {
	hospitalID := c.Param("hospitalId")
	authUserID := middleware.GetUserID(c)
	ctx := context.Background()

	var hospital models.Hospital
	if err := database.Col(database.ColHospitals).FindOne(ctx, bson.M{"_id": hospitalID}).Decode(&hospital); err != nil {
		utils.NotFoundResponse(c, "Hospital not found")
		return
	}
	role := middleware.GetUserRole(c)
	if hospital.UserID != authUserID && role != string(models.RoleAdmin) && role != string(models.RoleSuperAdmin) {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	cursor, _ := database.Col(database.ColEmergencies).Find(ctx, bson.M{
		"hospital_id": hospitalID,
		"status":      bson.M{"$nin": []string{string(models.EmergencyStatusResolved), string(models.EmergencyStatusCancelled)}},
	})
	var emergencies []models.Emergency
	cursor.All(ctx, &emergencies)
	utils.SuccessResponse(c, "Emergencies fetched", emergencies)
}
