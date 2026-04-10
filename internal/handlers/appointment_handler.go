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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateAppointment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	ctx := context.Background()

	var svc models.Service
	if err := database.Col(database.ColServices).FindOne(ctx,
		bson.M{"_id": req.ServiceID, "is_active": true}).Decode(&svc); err != nil {
		utils.NotFoundResponse(c, "Service not found")
		return
	}

	amount := svc.BasePrice
	if req.DoctorID != "" {
		var doc models.Doctor
		if err := database.Col(database.ColDoctors).FindOne(ctx,
			bson.M{"_id": req.DoctorID}).Decode(&doc); err == nil {
			if req.AppointmentType == models.AppointmentTypeHomeVisit {
				amount = doc.HomeVisitFee
			} else if req.AppointmentType == models.AppointmentTypeOnline {
				amount = doc.ConsultationFee
			}
		}
	}

	now := time.Now()
	appt := models.Appointment{
		ID:               uuid.New().String(),
		PatientUserID:    userID,
		DoctorID:         req.DoctorID,
		NurseID:          req.NurseID,
		ServiceID:        req.ServiceID,
		AppointmentType:  req.AppointmentType,
		Status:           models.AppointmentStatusPending,
		ScheduledAt:      req.ScheduledAt,
		Duration:         svc.Duration,
		PatientAddress:   req.PatientAddress,
		PatientLatitude:  req.PatientLatitude,
		PatientLongitude: req.PatientLongitude,
		Notes:            req.Notes,
		TotalAmount:      amount,
		PaymentStatus:    "PENDING",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	database.Col(database.ColAppointments).InsertOne(ctx, appt)
	createNotification(ctx, userID, "Appointment Booked",
		"Your appointment has been booked. You will receive a confirmation shortly.",
		"APPOINTMENT", appt.ID)
	utils.CreatedResponse(c, "Appointment created", appt)
}

func GetMyAppointments(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, limit := parsePage(c)
	statusFilter := c.Query("status")
	typeFilter := c.Query("type")
	ctx := context.Background()

	query := bson.M{"patient_user_id": userID}
	if statusFilter != "" {
		query["status"] = statusFilter
	}
	if typeFilter != "" {
		query["appointment_type"] = typeFilter
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

func GetAppointment(c *gin.Context) {
	appointmentID := c.Param("appointmentId")
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	ctx := context.Background()

	var appt models.Appointment
	if err := database.Col(database.ColAppointments).FindOne(ctx,
		bson.M{"_id": appointmentID}).Decode(&appt); err != nil {
		utils.NotFoundResponse(c, "Appointment not found")
		return
	}

	isAdmin := role == string(models.RoleAdmin) || role == string(models.RoleSuperAdmin)
	isPatient := appt.PatientUserID == userID
	isDoctor := false
	if appt.DoctorID != "" {
		var doc models.Doctor
		if err := database.Col(database.ColDoctors).FindOne(ctx,
			bson.M{"_id": appt.DoctorID}).Decode(&doc); err == nil {
			isDoctor = doc.UserID == userID
		}
	}
	if !isPatient && !isDoctor && !isAdmin {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}
	utils.SuccessResponse(c, "Appointment fetched", appt)
}

func UpdateAppointmentStatus(c *gin.Context) {
	appointmentID := c.Param("appointmentId")
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	ctx := context.Background()

	var appt models.Appointment
	if err := database.Col(database.ColAppointments).FindOne(ctx,
		bson.M{"_id": appointmentID}).Decode(&appt); err != nil {
		utils.NotFoundResponse(c, "Appointment not found")
		return
	}

	isAdmin := role == string(models.RoleAdmin) || role == string(models.RoleSuperAdmin)
	isDoctor := false
	if appt.DoctorID != "" {
		var doc models.Doctor
		if err := database.Col(database.ColDoctors).FindOne(ctx,
			bson.M{"_id": appt.DoctorID}).Decode(&doc); err == nil {
			isDoctor = doc.UserID == userID
		}
	}
	if !isDoctor && !isAdmin {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	var req models.UpdateAppointmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	set := bson.M{"status": req.Status, "updated_at": time.Now()}
	if req.ProviderNotes != "" {
		set["provider_notes"] = req.ProviderNotes
	}
	if req.CancelledReason != "" {
		set["cancelled_reason"] = req.CancelledReason
	}
	database.Col(database.ColAppointments).UpdateOne(ctx,
		bson.M{"_id": appointmentID}, bson.M{"$set": set})
	createNotification(ctx, appt.PatientUserID, "Appointment Update",
		"Your appointment status updated to: "+string(req.Status),
		"APPOINTMENT", appointmentID)
	utils.SuccessResponse(c, "Status updated", nil)
}

func CancelAppointment(c *gin.Context) {
	appointmentID := c.Param("appointmentId")
	userID := middleware.GetUserID(c)
	ctx := context.Background()

	var appt models.Appointment
	if err := database.Col(database.ColAppointments).FindOne(ctx,
		bson.M{"_id": appointmentID, "patient_user_id": userID}).Decode(&appt); err != nil {
		utils.NotFoundResponse(c, "Appointment not found")
		return
	}
	if appt.Status == models.AppointmentStatusCompleted ||
		appt.Status == models.AppointmentStatusCancelled {
		utils.BadRequestResponse(c, "Cannot cancel in current state")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)
	database.Col(database.ColAppointments).UpdateOne(ctx,
		bson.M{"_id": appointmentID},
		bson.M{"$set": bson.M{
			"status":           models.AppointmentStatusCancelled,
			"cancelled_reason": req.Reason,
			"updated_at":       time.Now(),
		}})
	utils.SuccessResponse(c, "Appointment cancelled", nil)
}

func GetPrescriptions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, _ := database.Col(database.ColPrescriptions).Find(context.Background(),
		bson.M{"patient_user_id": userID}, opts)
	var prescriptions []models.Prescription
	cursor.All(context.Background(), &prescriptions)
	utils.SuccessResponse(c, "Prescriptions fetched", prescriptions)
}

func GetFamilyAppointments(c *gin.Context) {
	familyUserID := middleware.GetUserID(c)
	patientID := c.Param("patientId")
	ctx := context.Background()

	var link models.FamilyMember
	if err := database.Col(database.ColFamilyMembers).FindOne(ctx,
		bson.M{"family_user_id": familyUserID, "patient_user_id": patientID}).Decode(&link); err != nil {
		utils.ForbiddenResponse(c, "Not authorized to view this patient's appointments")
		return
	}

	opts := options.Find().SetSort(bson.D{{Key: "scheduled_at", Value: -1}})
	cursor, _ := database.Col(database.ColAppointments).Find(ctx,
		bson.M{"patient_user_id": patientID}, opts)
	var appointments []models.Appointment
	cursor.All(ctx, &appointments)
	utils.SuccessResponse(c, "Patient appointments fetched", appointments)
}
