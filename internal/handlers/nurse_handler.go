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

func RegisterNurse(c *gin.Context) {
	var req models.RegisterNurseRequest
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

	now := time.Now()
	nurse := models.Nurse{
		ID:                uuid.New().String(),
		UserID:            req.UserID,
		HospitalID:        req.HospitalID,
		Qualification:     req.Qualification,
		Category:          req.Category,
		IDProofType:       req.IDProofType,
		IDProofNumber:     req.IDProofNumber,
		YearsOfExperience: req.YearsOfExperience,
		ServiceArea:       req.ServiceArea,
		ShiftAvailability: req.ShiftAvailability,
		EmergencyCapable:  req.EmergencyCapable,
		HourlyRate:        req.HourlyRate,
		DailyRate:         req.DailyRate,
		IsAvailable:       true,
		ApprovalStatus:    models.NurseApprovalPending,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if _, err := database.Col(database.ColNurses).InsertOne(ctx, nurse); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to register nurse")
		return
	}
	database.Col(database.ColUsers).UpdateOne(ctx, bson.M{"_id": req.UserID},
		bson.M{"$set": bson.M{"role": models.RoleNurse}})
	utils.CreatedResponse(c, "Nurse registered. Pending admin approval.", nurse)
}

func GetNurseProfile(c *gin.Context) {
	nurseID := c.Param("nurseId")
	var nurse models.Nurse
	err := database.Col(database.ColNurses).FindOne(context.Background(), bson.M{"_id": nurseID}).Decode(&nurse)
	if err == mongo.ErrNoDocuments {
		utils.NotFoundResponse(c, "Nurse not found")
		return
	}
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch nurse")
		return
	}
	utils.SuccessResponse(c, "Nurse profile fetched", buildNurseResponse(c.Request.Context(), &nurse))
}

func UpdateNurseAvailability(c *gin.Context) {
	nurseID := c.Param("nurseId")
	authUserID := middleware.GetUserID(c)
	ctx := context.Background()

	var nurse models.Nurse
	if err := database.Col(database.ColNurses).FindOne(ctx, bson.M{"_id": nurseID}).Decode(&nurse); err != nil {
		utils.NotFoundResponse(c, "Nurse not found")
		return
	}
	if nurse.UserID != authUserID {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	var req struct {
		IsAvailable bool `json:"is_available"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	database.Col(database.ColNurses).UpdateOne(ctx, bson.M{"_id": nurseID},
		bson.M{"$set": bson.M{"is_available": req.IsAvailable, "updated_at": time.Now()}})
	utils.SuccessResponse(c, "Availability updated", gin.H{"is_available": req.IsAvailable})
}

func SearchNurses(c *gin.Context) {
	category := c.Query("category")
	shiftAvail := c.Query("shift_availability")
	emergencyOnly := c.Query("emergency_capable")

	query := bson.M{"approval_status": models.NurseApprovalApproved, "is_available": true}
	if category != "" {
		query["category"] = category
	}
	if shiftAvail != "" {
		query["$or"] = bson.A{
			bson.M{"shift_availability": shiftAvail},
			bson.M{"shift_availability": "BOTH"},
		}
	}
	if emergencyOnly == "true" {
		query["emergency_capable"] = true
	}

	ctx := context.Background()
	total, _ := database.Col(database.ColNurses).CountDocuments(ctx, query)

	opts := options.Find().SetSort(bson.D{{Key: "rating", Value: -1}})
	cursor, _ := database.Col(database.ColNurses).Find(ctx, query, opts)
	var nurses []models.Nurse
	cursor.All(ctx, &nurses)

	responses := make([]models.NurseResponse, 0, len(nurses))
	for i := range nurses {
		responses = append(responses, buildNurseResponse(ctx, &nurses[i]))
	}
	utils.SuccessResponse(c, "Nurses fetched", gin.H{"data": responses, "total": total})
}

func GetNurseEarnings(c *gin.Context) {
	authUserID := middleware.GetUserID(c)
	ctx := context.Background()

	var nurse models.Nurse
	if err := database.Col(database.ColNurses).FindOne(ctx, bson.M{"user_id": authUserID}).Decode(&nurse); err != nil {
		utils.NotFoundResponse(c, "Nurse profile not found")
		return
	}

	completedCount, _ := database.Col(database.ColBookings).CountDocuments(ctx,
		bson.M{"professional_id": nurse.ID, "status": models.BookingStatusCompleted})

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"professional_id": nurse.ID, "status": models.BookingStatusCompleted}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "total": bson.M{"$sum": "$total_amount"}}}},
	}
	cursor, _ := database.Col(database.ColBookings).Aggregate(ctx, pipeline)
	var result []struct {
		Total float64 `bson:"total"`
	}
	cursor.All(ctx, &result)

	total := 0.0
	if len(result) > 0 {
		total = result[0].Total
	}
	utils.SuccessResponse(c, "Earnings fetched", gin.H{"total_earnings": total, "completed_count": completedCount})
}

func buildNurseResponse(ctx context.Context, n *models.Nurse) models.NurseResponse {
	resp := models.NurseResponse{
		NurseID:           n.ID,
		UserID:            n.UserID,
		Qualification:     n.Qualification,
		Category:          n.Category,
		YearsOfExperience: n.YearsOfExperience,
		HourlyRate:        n.HourlyRate,
		DailyRate:         n.DailyRate,
		ShiftAvailability: n.ShiftAvailability,
		EmergencyCapable:  n.EmergencyCapable,
		IsAvailable:       n.IsAvailable,
		Rating:            n.Rating,
		TotalReviews:      n.TotalReviews,
	}
	var user models.User
	if err := database.Col(database.ColUsers).FindOne(ctx, bson.M{"_id": n.UserID}).Decode(&user); err == nil {
		resp.Name = user.Name
		resp.ProfileImage = user.ProfileImage
	}
	if n.HospitalID != "" {
		var h models.Hospital
		if err := database.Col(database.ColHospitals).FindOne(ctx, bson.M{"_id": n.HospitalID}).Decode(&h); err == nil {
			resp.HospitalName = h.Name
		}
	}
	return resp
}
