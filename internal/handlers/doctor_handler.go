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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RegisterDoctor(c *gin.Context) {
	var req models.RegisterDoctorRequest
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

	// Duplicate license check
	count, _ := database.Col(database.ColDoctors).CountDocuments(ctx, bson.M{"license_number": req.LicenseNumber})
	if count > 0 {
		utils.ErrorResponse(c, 409, "License number already registered", "duplicate_license")
		return
	}

	now := time.Now()
	doctor := models.Doctor{
		ID:                uuid.New().String(),
		UserID:            req.UserID,
		HospitalID:        req.HospitalID,
		Qualification:     req.Qualification,
		Specialty:         req.Specialty,
		SubSpecialty:      req.SubSpecialty,
		LicenseNumber:     req.LicenseNumber,
		YearsOfExperience: req.YearsOfExperience,
		Languages:         req.Languages,
		ConsultationFee:   req.ConsultationFee,
		HomeVisitFee:      req.HomeVisitFee,
		ConsultationType:  req.ConsultationType,
		ServiceRadius:     req.ServiceRadius,
		Bio:               req.Bio,
		Availability:      models.DoctorAvailOffline,
		ApprovalStatus:    models.DoctorApprovalPending,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if _, err := database.Col(database.ColDoctors).InsertOne(ctx, doctor); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to register doctor")
		return
	}
	database.Col(database.ColUsers).UpdateOne(ctx, bson.M{"_id": req.UserID}, bson.M{"$set": bson.M{"role": models.RoleDoctor}})
	utils.CreatedResponse(c, "Doctor registered. Pending admin approval.", doctor)
}

func GetDoctorProfile(c *gin.Context) {
	doctorID := c.Param("doctorId")
	var doctor models.Doctor
	err := database.Col(database.ColDoctors).FindOne(context.Background(), bson.M{"_id": doctorID}).Decode(&doctor)
	if err == mongo.ErrNoDocuments {
		utils.NotFoundResponse(c, "Doctor not found")
		return
	}
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch doctor")
		return
	}
	resp := buildDoctorResponse(c.Request.Context(), &doctor)
	utils.SuccessResponse(c, "Doctor profile fetched", resp)
}

func UpdateDoctorProfile(c *gin.Context) {
	doctorID := c.Param("doctorId")
	authUserID := middleware.GetUserID(c)
	ctx := context.Background()

	var doctor models.Doctor
	if err := database.Col(database.ColDoctors).FindOne(ctx, bson.M{"_id": doctorID}).Decode(&doctor); err != nil {
		utils.NotFoundResponse(c, "Doctor not found")
		return
	}
	if doctor.UserID != authUserID {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	var req models.UpdateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	set := bson.M{"updated_at": time.Now()}
	if req.Specialty != "" {
		set["specialty"] = req.Specialty
	}
	if req.SubSpecialty != "" {
		set["sub_specialty"] = req.SubSpecialty
	}
	if req.ConsultationFee > 0 {
		set["consultation_fee"] = req.ConsultationFee
	}
	if req.HomeVisitFee > 0 {
		set["home_visit_fee"] = req.HomeVisitFee
	}
	if req.ConsultationType != "" {
		set["consultation_type"] = req.ConsultationType
	}
	if req.ServiceRadius > 0 {
		set["service_radius"] = req.ServiceRadius
	}
	if req.Availability != "" {
		set["availability"] = req.Availability
	}
	if req.Bio != "" {
		set["bio"] = req.Bio
	}
	if req.Languages != "" {
		set["languages"] = req.Languages
	}

	database.Col(database.ColDoctors).UpdateOne(ctx, bson.M{"_id": doctorID}, bson.M{"$set": set})
	database.Col(database.ColDoctors).FindOne(ctx, bson.M{"_id": doctorID}).Decode(&doctor)
	utils.SuccessResponse(c, "Doctor updated", buildDoctorResponse(ctx, &doctor))
}

func SetDoctorAvailability(c *gin.Context) {
	doctorID := c.Param("doctorId")
	authUserID := middleware.GetUserID(c)
	ctx := context.Background()

	var doctor models.Doctor
	if err := database.Col(database.ColDoctors).FindOne(ctx, bson.M{"_id": doctorID}).Decode(&doctor); err != nil {
		utils.NotFoundResponse(c, "Doctor not found")
		return
	}
	if doctor.UserID != authUserID {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	var req struct {
		Availability models.DoctorAvailability `json:"availability" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	database.Col(database.ColDoctors).UpdateOne(ctx, bson.M{"_id": doctorID},
		bson.M{"$set": bson.M{"availability": req.Availability, "updated_at": time.Now()}})
	utils.SuccessResponse(c, "Availability updated", gin.H{"availability": req.Availability})
}

func SearchDoctors(c *gin.Context) {
	var filter models.DoctorSearchFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	query := bson.M{"approval_status": models.DoctorApprovalApproved}
	if filter.Specialty != "" {
		query["specialty"] = bson.M{"$regex": filter.Specialty, "$options": "i"}
	}
	if filter.MinRating > 0 {
		query["rating"] = bson.M{"$gte": filter.MinRating}
	}
	if filter.MaxFee > 0 {
		query["consultation_fee"] = bson.M{"$lte": filter.MaxFee}
	}
	if filter.Language != "" {
		query["languages"] = bson.M{"$regex": filter.Language, "$options": "i"}
	}
	if filter.ConsultationType != "" {
		query["$or"] = bson.A{
			bson.M{"consultation_type": filter.ConsultationType},
			bson.M{"consultation_type": "BOTH"},
		}
	}

	ctx := context.Background()
	total, _ := database.Col(database.ColDoctors).CountDocuments(ctx, query)

	opts := options.Find().
		SetSort(bson.D{{Key: "rating", Value: -1}, {Key: "years_of_experience", Value: -1}}).
		SetSkip(int64((filter.Page - 1) * filter.Limit)).
		SetLimit(int64(filter.Limit))

	cursor, _ := database.Col(database.ColDoctors).Find(ctx, query, opts)
	var doctors []models.Doctor
	cursor.All(ctx, &doctors)

	responses := make([]models.DoctorResponse, 0, len(doctors))
	for i := range doctors {
		responses = append(responses, buildDoctorResponse(ctx, &doctors[i]))
	}
	utils.PaginatedSuccessResponse(c, "Doctors fetched", responses, filter.Page, filter.Limit, total)
}

func GetDoctorSchedule(c *gin.Context) {
	doctorID := c.Param("doctorId")
	var doctor models.Doctor
	if err := database.Col(database.ColDoctors).FindOne(context.Background(), bson.M{"_id": doctorID}).Decode(&doctor); err != nil {
		utils.NotFoundResponse(c, "Doctor not found")
		return
	}
	utils.SuccessResponse(c, "Schedule fetched", doctor.Schedules)
}

func SetDoctorSchedule(c *gin.Context) {
	doctorID := c.Param("doctorId")
	authUserID := middleware.GetUserID(c)
	ctx := context.Background()

	var doctor models.Doctor
	if err := database.Col(database.ColDoctors).FindOne(ctx, bson.M{"_id": doctorID}).Decode(&doctor); err != nil {
		utils.NotFoundResponse(c, "Doctor not found")
		return
	}
	if doctor.UserID != authUserID {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	var schedules []models.DoctorSchedule
	if err := c.ShouldBindJSON(&schedules); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	database.Col(database.ColDoctors).UpdateOne(ctx, bson.M{"_id": doctorID},
		bson.M{"$set": bson.M{"schedules": schedules, "updated_at": time.Now()}})
	utils.SuccessResponse(c, "Schedule updated", schedules)
}

func GetMyDoctorAppointments(c *gin.Context) {
	authUserID := middleware.GetUserID(c)
	ctx := context.Background()

	var doctor models.Doctor
	if err := database.Col(database.ColDoctors).FindOne(ctx, bson.M{"user_id": authUserID}).Decode(&doctor); err != nil {
		utils.NotFoundResponse(c, "Doctor profile not found")
		return
	}

	page, limit := parsePage(c)
	statusFilter := c.Query("status")

	query := bson.M{"doctor_id": doctor.ID}
	if statusFilter != "" {
		query["status"] = statusFilter
	}

	total, _ := database.Col(database.ColAppointments).CountDocuments(ctx, query)
	opts := options.Find().
		SetSort(bson.D{{Key: "scheduled_at", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, _ := database.Col(database.ColAppointments).Find(ctx, query, opts)
	var appointments []models.Appointment
	cursor.All(ctx, &appointments)

	utils.PaginatedSuccessResponse(c, "Appointments fetched", appointments, page, limit, total)
}

func UploadPrescription(c *gin.Context) {
	authUserID := middleware.GetUserID(c)
	ctx := context.Background()

	var doctor models.Doctor
	if err := database.Col(database.ColDoctors).FindOne(ctx, bson.M{"user_id": authUserID}).Decode(&doctor); err != nil {
		utils.NotFoundResponse(c, "Doctor profile not found")
		return
	}

	var req struct {
		AppointmentID string  `json:"appointment_id" binding:"required"`
		PatientUserID string  `json:"patient_user_id" binding:"required"`
		Medications   string  `json:"medications"    binding:"required"`
		Instructions  string  `json:"instructions"`
		FollowUpDate  *string `json:"follow_up_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	now := time.Now()
	prescription := models.Prescription{
		ID:            uuid.New().String(),
		AppointmentID: req.AppointmentID,
		DoctorID:      doctor.ID,
		PatientUserID: req.PatientUserID,
		Medications:   req.Medications,
		Instructions:  req.Instructions,
		IssuedAt:      now,
		CreatedAt:     now,
	}
	database.Col(database.ColPrescriptions).InsertOne(ctx, prescription)
	utils.CreatedResponse(c, "Prescription uploaded", prescription)
}

func GetDoctorEarnings(c *gin.Context) {
	authUserID := middleware.GetUserID(c)
	ctx := context.Background()

	var doctor models.Doctor
	if err := database.Col(database.ColDoctors).FindOne(ctx, bson.M{"user_id": authUserID}).Decode(&doctor); err != nil {
		utils.NotFoundResponse(c, "Doctor profile not found")
		return
	}

	completedCount, _ := database.Col(database.ColAppointments).CountDocuments(ctx,
		bson.M{"doctor_id": doctor.ID, "status": models.AppointmentStatusCompleted, "payment_status": "PAID"})

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"doctor_id": doctor.ID, "status": models.AppointmentStatusCompleted, "payment_status": "PAID"}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "total": bson.M{"$sum": "$total_amount"}}}},
	}
	cursor, _ := database.Col(database.ColAppointments).Aggregate(ctx, pipeline)
	var result []struct {
		Total float64 `bson:"total"`
	}
	cursor.All(ctx, &result)

	total := 0.0
	if len(result) > 0 {
		total = result[0].Total
	}

	utils.SuccessResponse(c, "Earnings fetched", gin.H{
		"total_earnings":  total,
		"completed_count": completedCount,
	})
}

// buildDoctorResponse enriches a doctor with its user's name and profile image
func buildDoctorResponse(ctx context.Context, d *models.Doctor) models.DoctorResponse {
	resp := models.DoctorResponse{
		DoctorID:          d.ID,
		UserID:            d.UserID,
		Qualification:     d.Qualification,
		Specialty:         d.Specialty,
		SubSpecialty:      d.SubSpecialty,
		YearsOfExperience: d.YearsOfExperience,
		ConsultationFee:   d.ConsultationFee,
		HomeVisitFee:      d.HomeVisitFee,
		ConsultationType:  d.ConsultationType,
		Availability:      d.Availability,
		Rating:            d.Rating,
		TotalReviews:      d.TotalReviews,
		Languages:         d.Languages,
		Bio:               d.Bio,
		ServiceRadius:     d.ServiceRadius,
	}
	var user models.User
	if err := database.Col(database.ColUsers).FindOne(ctx, bson.M{"_id": d.UserID}).Decode(&user); err == nil {
		resp.Name = user.Name
		resp.ProfileImage = user.ProfileImage
	}
	if d.HospitalID != "" {
		var h models.Hospital
		if err := database.Col(database.ColHospitals).FindOne(ctx, bson.M{"_id": d.HospitalID}).Decode(&h); err == nil {
			resp.HospitalName = h.Name
		}
	}
	return resp
}
