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

// TriggerSOS — patient presses emergency button
func TriggerSOS(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.SOSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	ctx := context.Background()

	now := time.Now()
	emergency := models.Emergency{
		ID:                 uuid.New().String(),
		PatientUserID:      userID,
		Status:             models.EmergencyStatusTriggered,
		Priority:           models.EmergencyPriorityHigh,
		PatientLatitude:    req.PatientLatitude,
		PatientLongitude:   req.PatientLongitude,
		PatientAddress:     req.PatientAddress,
		SymptomDescription: req.SymptomDescription,
		EmergencyType:      req.EmergencyType,
		HospitalNotified:   false,
		FamilyNotified:     false,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	database.Col(database.ColEmergencies).InsertOne(ctx, emergency)

	// Notify family members
	cursor, _ := database.Col(database.ColFamilyMembers).Find(ctx,
		bson.M{"patient_user_id": userID, "alerts_enabled": true})
	var family []models.FamilyMember
	cursor.All(ctx, &family)
	for _, f := range family {
		createNotification(ctx, f.FamilyUserID,
			"🚨 Emergency Alert",
			"Your family member has triggered an emergency SOS. Please check immediately.",
			"EMERGENCY", emergency.ID)
	}

	database.Col(database.ColEmergencies).UpdateOne(ctx, bson.M{"_id": emergency.ID},
		bson.M{"$set": bson.M{"family_notified": len(family) > 0}})

	utils.CreatedResponse(c, "Emergency triggered. Help is on the way.", emergency)
}

// GetEmergency returns a single emergency record
func GetEmergency(c *gin.Context) {
	emergencyID := c.Param("emergencyId")
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	ctx := context.Background()

	var emergency models.Emergency
	if err := database.Col(database.ColEmergencies).FindOne(ctx,
		bson.M{"_id": emergencyID}).Decode(&emergency); err != nil {
		utils.NotFoundResponse(c, "Emergency not found")
		return
	}

	isAdmin := role == string(models.RoleAdmin) || role == string(models.RoleSuperAdmin)
	isHospital := role == string(models.RoleHospital)
	if emergency.PatientUserID != userID && !isAdmin && !isHospital {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}
	utils.SuccessResponse(c, "Emergency fetched", emergency)
}

// GetMyEmergencies returns all emergencies for the logged-in patient
func GetMyEmergencies(c *gin.Context) {
	userID := middleware.GetUserID(c)
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, _ := database.Col(database.ColEmergencies).Find(context.Background(),
		bson.M{"patient_user_id": userID}, opts)
	var emergencies []models.Emergency
	cursor.All(context.Background(), &emergencies)
	utils.SuccessResponse(c, "Emergencies fetched", emergencies)
}

// UpdateEmergencyStatus — admin/hospital updates emergency progress
func UpdateEmergencyStatus(c *gin.Context) {
	emergencyID := c.Param("emergencyId")
	role := middleware.GetUserRole(c)

	isAdmin := role == string(models.RoleAdmin) || role == string(models.RoleSuperAdmin)
	isHospital := role == string(models.RoleHospital)
	if !isAdmin && !isHospital {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	var req models.UpdateEmergencyStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	ctx := context.Background()
	set := bson.M{"status": req.Status, "updated_at": time.Now()}
	if req.Priority != "" {
		set["priority"] = req.Priority
	}
	if req.Notes != "" {
		set["notes"] = req.Notes
	}
	if req.AmbulanceID != "" {
		set["ambulance_id"] = req.AmbulanceID
	}
	if req.HospitalID != "" {
		set["hospital_id"] = req.HospitalID
		set["hospital_notified"] = true
	}

	now := time.Now()
	switch req.Status {
	case models.EmergencyStatusDispatched:
		set["dispatched_at"] = now
	case models.EmergencyStatusArrived:
		set["arrived_at"] = now
	case models.EmergencyStatusResolved:
		set["resolved_at"] = now
	}

	database.Col(database.ColEmergencies).UpdateOne(ctx,
		bson.M{"_id": emergencyID}, bson.M{"$set": set})

	// Notify patient
	var emergency models.Emergency
	database.Col(database.ColEmergencies).FindOne(ctx, bson.M{"_id": emergencyID}).Decode(&emergency)
	if emergency.PatientUserID != "" {
		createNotification(ctx, emergency.PatientUserID,
			"Emergency Update", "Your emergency status: "+string(req.Status),
			"EMERGENCY", emergencyID)
	}

	utils.SuccessResponse(c, "Emergency status updated", nil)
}

// GetActiveEmergencies — admin view of all active emergencies
func GetActiveEmergencies(c *gin.Context) {
	role := middleware.GetUserRole(c)
	isAdmin := role == string(models.RoleAdmin) || role == string(models.RoleSuperAdmin)
	isHospital := role == string(models.RoleHospital)
	if !isAdmin && !isHospital {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	ctx := context.Background()
	query := bson.M{
		"status": bson.M{"$nin": []string{
			string(models.EmergencyStatusResolved),
			string(models.EmergencyStatusCancelled),
		}},
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, _ := database.Col(database.ColEmergencies).Find(ctx, query, opts)
	var emergencies []models.Emergency
	cursor.All(ctx, &emergencies)
	utils.SuccessResponse(c, "Active emergencies", emergencies)
}

// ---- Medical Records ----

func UploadMedicalRecord(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		RecordType   string `json:"record_type"   binding:"required"`
		Title        string `json:"title"         binding:"required"`
		Description  string `json:"description"`
		FileURL      string `json:"file_url"`
		RecordDate   string `json:"record_date"   binding:"required"`
		DoctorName   string `json:"doctor_name"`
		HospitalName string `json:"hospital_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	recordDate, err := time.Parse("2006-01-02", req.RecordDate)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid record_date format. Use YYYY-MM-DD")
		return
	}

	now := time.Now()
	record := models.MedicalRecord{
		ID:           uuid.New().String(),
		UserID:       userID,
		RecordType:   req.RecordType,
		Title:        req.Title,
		Description:  req.Description,
		FileURL:      req.FileURL,
		RecordDate:   recordDate,
		DoctorName:   req.DoctorName,
		HospitalName: req.HospitalName,
		IsShared:     false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	database.Col(database.ColMedicalRecords).InsertOne(context.Background(), record)
	utils.CreatedResponse(c, "Medical record uploaded", record)
}

func GetMyMedicalRecords(c *gin.Context) {
	userID := middleware.GetUserID(c)
	recordType := c.Query("type")

	query := bson.M{"user_id": userID}
	if recordType != "" {
		query["record_type"] = recordType
	}

	opts := options.Find().SetSort(bson.D{{Key: "record_date", Value: -1}})
	cursor, _ := database.Col(database.ColMedicalRecords).Find(context.Background(), query, opts)
	var records []models.MedicalRecord
	cursor.All(context.Background(), &records)
	utils.SuccessResponse(c, "Medical records fetched", records)
}

func DeleteMedicalRecord(c *gin.Context) {
	userID := middleware.GetUserID(c)
	recordID := c.Param("recordId")

	res, _ := database.Col(database.ColMedicalRecords).DeleteOne(context.Background(),
		bson.M{"_id": recordID, "user_id": userID})
	if res.DeletedCount == 0 {
		utils.NotFoundResponse(c, "Record not found")
		return
	}
	utils.SuccessResponse(c, "Record deleted", nil)
}

func GetFamilyMedicalRecords(c *gin.Context) {
	familyUserID := middleware.GetUserID(c)
	patientID := c.Param("patientId")
	ctx := context.Background()

	var link models.FamilyMember
	if err := database.Col(database.ColFamilyMembers).FindOne(ctx,
		bson.M{"family_user_id": familyUserID, "patient_user_id": patientID}).Decode(&link); err != nil {
		utils.ForbiddenResponse(c, "Not authorized")
		return
	}

	opts := options.Find().SetSort(bson.D{{Key: "record_date", Value: -1}})
	cursor, _ := database.Col(database.ColMedicalRecords).Find(ctx,
		bson.M{"user_id": patientID, "is_shared": true}, opts)
	var records []models.MedicalRecord
	cursor.All(ctx, &records)
	utils.SuccessResponse(c, "Patient records fetched", records)
}
