package handlers

import (
	"context"
	"time"

	"home_care_backend/internal/database"
	"home_care_backend/internal/models"
	"home_care_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetServiceCategories(c *gin.Context) {
	cursor, _ := database.Col(database.ColCategories).Find(context.Background(),
		bson.M{"is_active": true})
	var cats []models.ServiceCategory
	cursor.All(context.Background(), &cats)
	utils.SuccessResponse(c, "Categories fetched", cats)
}

func GetAllServices(c *gin.Context) {
	query := bson.M{"is_active": true}
	if c.Query("is_quick") == "true" {
		query["is_quick"] = true
	}
	if c.Query("is_emergency") == "true" {
		query["is_emergency"] = true
	}
	cursor, _ := database.Col(database.ColServices).Find(context.Background(), query)
	var services []models.Service
	cursor.All(context.Background(), &services)
	utils.SuccessResponse(c, "Services fetched", services)
}

func GetCategoryServices(c *gin.Context) {
	cursor, _ := database.Col(database.ColServices).Find(context.Background(),
		bson.M{"category_id": c.Param("categoryId"), "is_active": true})
	var services []models.Service
	cursor.All(context.Background(), &services)
	utils.SuccessResponse(c, "Services fetched", services)
}

func GetService(c *gin.Context) {
	var svc models.Service
	err := database.Col(database.ColServices).FindOne(context.Background(),
		bson.M{"_id": c.Param("serviceId")}).Decode(&svc)
	if err == mongo.ErrNoDocuments {
		utils.NotFoundResponse(c, "Service not found")
		return
	}
	utils.SuccessResponse(c, "Service fetched", svc)
}

func GetProfessionals(c *gin.Context) {
	query := bson.M{}
	if c.DefaultQuery("available_only", "true") == "true" {
		query["is_available"] = true
	}
	if s := c.Query("service_name"); s != "" {
		query["$or"] = bson.A{
			bson.M{"service_name": bson.M{"$regex": s, "$options": "i"}},
			bson.M{"role": bson.M{"$regex": s, "$options": "i"}},
		}
	}
	if zoneID := c.Query("zone_id"); zoneID != "" {
		query["zone_id"] = zoneID
	}

	ctx := context.Background()
	cursor, _ := database.Col(database.ColProfessionals).Find(ctx, query)
	var pros []models.Professional
	cursor.All(ctx, &pros)

	// Pre-fetch zone names for efficiency
	zoneNames := map[string]string{}

	responses := make([]models.ProfessionalResponse, 0, len(pros))
	for _, p := range pros {
		r := models.ProfessionalResponse{
			ID:                 p.ID,
			Role:               p.Role,
			ServiceName:        p.ServiceName,
			ZoneID:             p.ZoneID,
			Bio:                p.Bio,
			Qualification:      p.Qualification,
			Rating:             p.Rating,
			Available:          p.IsAvailable,
			YearsExperience:    p.YearsOfExperience,
			EstimatedDuration:  p.EstimatedDuration,
			HourlyRate:         p.HourlyRate,
			AvailableTimeStart: p.AvailableTimeStart,
			AvailableTimeEnd:   p.AvailableTimeEnd,
		}
		var u models.User
		if err := database.Col(database.ColUsers).FindOne(ctx, bson.M{"_id": p.UserID}).Decode(&u); err == nil {
			r.Name = u.Name
			r.ImageURL = u.ProfileImage
		}
		if p.ZoneID != "" {
			if name, ok := zoneNames[p.ZoneID]; ok {
				r.ZoneName = name
			} else {
				var z models.ServiceZone
				if err := database.Col(database.ColServiceZones).FindOne(ctx, bson.M{"_id": p.ZoneID}).Decode(&z); err == nil {
					zoneNames[p.ZoneID] = z.Name
					r.ZoneName = z.Name
				}
			}
		}
		responses = append(responses, r)
	}
	utils.SuccessResponse(c, "Professionals fetched", responses)
}

func GetProfessional(c *gin.Context) {
	ctx := context.Background()
	var p models.Professional
	err := database.Col(database.ColProfessionals).FindOne(ctx,
		bson.M{"_id": c.Param("professionalId")}).Decode(&p)
	if err == mongo.ErrNoDocuments {
		utils.NotFoundResponse(c, "Professional not found")
		return
	}
	r := models.ProfessionalResponse{
		ID:                 p.ID,
		Role:               p.Role,
		ServiceName:        p.ServiceName,
		Rating:             p.Rating,
		Available:          p.IsAvailable,
		YearsExperience:    p.YearsOfExperience,
		EstimatedDuration:  p.EstimatedDuration,
		AvailableTimeStart: p.AvailableTimeStart,
		AvailableTimeEnd:   p.AvailableTimeEnd,
	}
	var u models.User
	if err := database.Col(database.ColUsers).FindOne(ctx, bson.M{"_id": p.UserID}).Decode(&u); err == nil {
		r.Name = u.Name
		r.ImageURL = u.ProfileImage
	}
	utils.SuccessResponse(c, "Professional fetched", r)
}

func SubmitReview(c *gin.Context) {
	userID := getAuthUserID(c)
	var req struct {
		BookingID string `json:"booking_id"`
		DoctorID  string `json:"doctor_id"`
		NurseID   string `json:"nurse_id"`
		Rating    int    `json:"rating" binding:"required,min=1,max=5"`
		Comment   string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	ctx := context.Background()
	review := models.Review{
		ID:        uuid.New().String(),
		UserID:    userID,
		BookingID: req.BookingID,
		DoctorID:  req.DoctorID,
		NurseID:   req.NurseID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		IsPublic:  true,
		CreatedAt: time.Now(),
	}
	database.Col(database.ColReviews).InsertOne(ctx, review)

	if req.DoctorID != "" {
		updateProviderRating(ctx, database.ColDoctors, req.DoctorID, "doctor_id")
	}
	if req.NurseID != "" {
		updateProviderRating(ctx, database.ColNurses, req.NurseID, "nurse_id")
	}
	utils.CreatedResponse(c, "Review submitted", review)
}

func updateProviderRating(ctx context.Context, col, id, field string) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{field: id}}},
		{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"avg":   bson.M{"$avg": "$rating"},
			"count": bson.M{"$sum": 1},
		}}},
	}
	cursor, err := database.Col(database.ColReviews).Aggregate(ctx, pipeline)
	if err != nil {
		return
	}
	var result []struct {
		Avg   float64 `bson:"avg"`
		Count int64   `bson:"count"`
	}
	cursor.All(ctx, &result)
	if len(result) > 0 {
		database.Col(col).UpdateOne(ctx, bson.M{"_id": id},
			bson.M{"$set": bson.M{"rating": result[0].Avg, "total_reviews": result[0].Count}})
	}
}

func getAuthUserID(c *gin.Context) string {
	v, _ := c.Get("user_id")
	if id, ok := v.(string); ok {
		return id
	}
	return ""
}
