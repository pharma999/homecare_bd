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

// ---- User Management ----

func AdminListUsers(c *gin.Context) {
	page, limit := parsePage(c)
	role := c.Query("role")
	status := c.Query("status")
	ctx := context.Background()

	query := bson.M{}
	if role != "" {
		query["role"] = role
	}
	if status != "" {
		query["status"] = status
	}

	total, _ := database.Col(database.ColUsers).CountDocuments(ctx, query)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, _ := database.Col(database.ColUsers).Find(ctx, query, opts)
	var users []models.User
	cursor.All(ctx, &users)
	utils.PaginatedSuccessResponse(c, "Users fetched", users, page, limit, total)
}

func AdminGetUser(c *gin.Context) {
	var user models.User
	err := database.Col(database.ColUsers).FindOne(context.Background(),
		bson.M{"_id": c.Param("userId")}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		utils.NotFoundResponse(c, "User not found")
		return
	}
	utils.SuccessResponse(c, "User fetched", buildUserResponse(&user))
}

func AdminBlockUser(c *gin.Context) {
	userID := c.Param("userId")
	var req struct {
		Block bool `json:"block"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	blockStatus := models.BlockStatusUnblocked
	if req.Block {
		blockStatus = models.BlockStatusBlocked
	}
	database.Col(database.ColUsers).UpdateOne(context.Background(),
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"block_status": blockStatus, "updated_at": time.Now()}})
	utils.SuccessResponse(c, "User block status updated", gin.H{"block_status": blockStatus})
}

// ---- Hospital Approval ----

func AdminListHospitals(c *gin.Context) {
	page, limit := parsePage(c)
	approvalStatus := c.Query("approval_status")
	ctx := context.Background()

	query := bson.M{}
	if approvalStatus != "" {
		query["approval_status"] = approvalStatus
	}

	total, _ := database.Col(database.ColHospitals).CountDocuments(ctx, query)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, _ := database.Col(database.ColHospitals).Find(ctx, query, opts)
	var hospitals []models.Hospital
	cursor.All(ctx, &hospitals)
	utils.PaginatedSuccessResponse(c, "Hospitals fetched", hospitals, page, limit, total)
}

func AdminApproveHospital(c *gin.Context) {
	hospitalID := c.Param("hospitalId")
	var req struct {
		Status models.HospitalApprovalStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	database.Col(database.ColHospitals).UpdateOne(context.Background(),
		bson.M{"_id": hospitalID},
		bson.M{"$set": bson.M{"approval_status": req.Status, "updated_at": time.Now()}})
	utils.SuccessResponse(c, "Hospital approval updated", gin.H{"status": req.Status})
}

// ---- Doctor Approval ----

func AdminListDoctors(c *gin.Context) {
	page, limit := parsePage(c)
	approvalStatus := c.Query("approval_status")
	ctx := context.Background()

	query := bson.M{}
	if approvalStatus != "" {
		query["approval_status"] = approvalStatus
	}

	total, _ := database.Col(database.ColDoctors).CountDocuments(ctx, query)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, _ := database.Col(database.ColDoctors).Find(ctx, query, opts)
	var doctors []models.Doctor
	cursor.All(ctx, &doctors)

	responses := make([]models.DoctorResponse, 0, len(doctors))
	for i := range doctors {
		responses = append(responses, buildDoctorResponse(ctx, &doctors[i]))
	}
	utils.PaginatedSuccessResponse(c, "Doctors fetched", responses, page, limit, total)
}

func AdminApproveDoctor(c *gin.Context) {
	doctorID := c.Param("doctorId")
	var req struct {
		Status models.DoctorApprovalStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	database.Col(database.ColDoctors).UpdateOne(context.Background(),
		bson.M{"_id": doctorID},
		bson.M{"$set": bson.M{"approval_status": req.Status, "updated_at": time.Now()}})
	utils.SuccessResponse(c, "Doctor approval updated", gin.H{"status": req.Status})
}

// ---- Nurse Approval ----

func AdminListNurses(c *gin.Context) {
	page, limit := parsePage(c)
	approvalStatus := c.Query("approval_status")
	ctx := context.Background()

	query := bson.M{}
	if approvalStatus != "" {
		query["approval_status"] = approvalStatus
	}

	total, _ := database.Col(database.ColNurses).CountDocuments(ctx, query)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, _ := database.Col(database.ColNurses).Find(ctx, query, opts)
	var nurses []models.Nurse
	cursor.All(ctx, &nurses)

	responses := make([]models.NurseResponse, 0, len(nurses))
	for i := range nurses {
		responses = append(responses, buildNurseResponse(ctx, &nurses[i]))
	}
	utils.PaginatedSuccessResponse(c, "Nurses fetched", responses, page, limit, total)
}

func AdminApproveNurse(c *gin.Context) {
	nurseID := c.Param("nurseId")
	var req struct {
		Status models.NurseApprovalStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	database.Col(database.ColNurses).UpdateOne(context.Background(),
		bson.M{"_id": nurseID},
		bson.M{"$set": bson.M{"approval_status": req.Status, "updated_at": time.Now()}})
	utils.SuccessResponse(c, "Nurse approval updated", gin.H{"status": req.Status})
}

// ---- Service Management ----

func AdminCreateService(c *gin.Context) {
	var req struct {
		CategoryID   string  `json:"category_id"   binding:"required"`
		Title        string  `json:"title"         binding:"required"`
		Slug         string  `json:"slug"          binding:"required"`
		Description  string  `json:"description"`
		Icon         string  `json:"icon"`
		Color        string  `json:"color"`
		BasePrice    float64 `json:"base_price"    binding:"required"`
		Duration     int     `json:"duration"`
		IsQuick      bool    `json:"is_quick"`
		IsEmergency  bool    `json:"is_emergency"`
		DisplayOrder int     `json:"display_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	now := time.Now()
	svc := models.Service{
		ID:           uuid.New().String(),
		CategoryID:   req.CategoryID,
		Title:        req.Title,
		Slug:         req.Slug,
		Description:  req.Description,
		Icon:         req.Icon,
		Color:        req.Color,
		BasePrice:    req.BasePrice,
		Duration:     req.Duration,
		IsActive:     true,
		IsQuick:      req.IsQuick,
		IsEmergency:  req.IsEmergency,
		DisplayOrder: req.DisplayOrder,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	database.Col(database.ColServices).InsertOne(context.Background(), svc)
	utils.CreatedResponse(c, "Service created", svc)
}

func AdminUpdateService(c *gin.Context) {
	serviceID := c.Param("serviceId")
	var req struct {
		Title        string  `json:"title"`
		Description  string  `json:"description"`
		BasePrice    float64 `json:"base_price"`
		Duration     int     `json:"duration"`
		IsActive     *bool   `json:"is_active"`
		IsQuick      *bool   `json:"is_quick"`
		DisplayOrder int     `json:"display_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	set := bson.M{"updated_at": time.Now()}
	if req.Title != "" {
		set["title"] = req.Title
	}
	if req.Description != "" {
		set["description"] = req.Description
	}
	if req.BasePrice > 0 {
		set["base_price"] = req.BasePrice
	}
	if req.Duration > 0 {
		set["duration"] = req.Duration
	}
	if req.IsActive != nil {
		set["is_active"] = *req.IsActive
	}
	if req.IsQuick != nil {
		set["is_quick"] = *req.IsQuick
	}
	database.Col(database.ColServices).UpdateOne(context.Background(),
		bson.M{"_id": serviceID}, bson.M{"$set": set})
	utils.SuccessResponse(c, "Service updated", nil)
}

func AdminCreateCategory(c *gin.Context) {
	var req struct {
		Name         string `json:"name"          binding:"required"`
		Slug         string `json:"slug"          binding:"required"`
		Description  string `json:"description"`
		Icon         string `json:"icon"`
		Color        string `json:"color"`
		DisplayOrder int    `json:"display_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	now := time.Now()
	cat := models.ServiceCategory{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Slug:         req.Slug,
		Description:  req.Description,
		Icon:         req.Icon,
		Color:        req.Color,
		IsActive:     true,
		DisplayOrder: req.DisplayOrder,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	database.Col(database.ColCategories).InsertOne(context.Background(), cat)
	utils.CreatedResponse(c, "Category created", cat)
}

// ---- Analytics ----

func AdminAnalytics(c *gin.Context) {
	role := middleware.GetUserRole(c)
	if role != string(models.RoleAdmin) && role != string(models.RoleSuperAdmin) {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}
	ctx := c.Request.Context()

	totalUsers, _ := database.Col(database.ColUsers).CountDocuments(ctx, bson.M{})
	totalDoctors, _ := database.Col(database.ColDoctors).CountDocuments(ctx,
		bson.M{"approval_status": models.DoctorApprovalApproved})
	totalNurses, _ := database.Col(database.ColNurses).CountDocuments(ctx,
		bson.M{"approval_status": models.NurseApprovalApproved})
	totalHospitals, _ := database.Col(database.ColHospitals).CountDocuments(ctx,
		bson.M{"approval_status": models.HospitalApprovalApproved})
	totalBookings, _ := database.Col(database.ColBookings).CountDocuments(ctx, bson.M{})
	totalAppointments, _ := database.Col(database.ColAppointments).CountDocuments(ctx, bson.M{})
	activeEmergencies, _ := database.Col(database.ColEmergencies).CountDocuments(ctx,
		bson.M{"status": bson.M{"$nin": []string{
			string(models.EmergencyStatusResolved),
			string(models.EmergencyStatusCancelled),
		}}})

	// Pending approvals (doctors + nurses + hospitals)
	pendingDoctors, _ := database.Col(database.ColDoctors).CountDocuments(ctx,
		bson.M{"approval_status": models.DoctorApprovalPending})
	pendingNurses, _ := database.Col(database.ColNurses).CountDocuments(ctx,
		bson.M{"approval_status": models.NurseApprovalPending})
	pendingHospitals, _ := database.Col(database.ColHospitals).CountDocuments(ctx,
		bson.M{"approval_status": models.HospitalApprovalPending})
	pendingApprovals := pendingDoctors + pendingNurses + pendingHospitals

	// Open support tickets
	openTickets, _ := database.Col(database.ColSupportTickets).CountDocuments(ctx,
		bson.M{"status": bson.M{"$in": []string{
			string(models.TicketOpen), string(models.TicketInProgress),
		}}})

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"status": "COMPLETED"}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "total": bson.M{"$sum": "$amount"}}}},
	}
	cursor, _ := database.Col(database.ColPayments).Aggregate(ctx, pipeline)
	var revenueResult []struct {
		Total float64 `bson:"total"`
	}
	cursor.All(ctx, &revenueResult)
	totalRevenue := 0.0
	if len(revenueResult) > 0 {
		totalRevenue = revenueResult[0].Total
	}

	utils.SuccessResponse(c, "Analytics fetched", gin.H{
		"total_users":        totalUsers,
		"total_doctors":      totalDoctors,
		"total_nurses":       totalNurses,
		"total_hospitals":    totalHospitals,
		"total_bookings":     totalBookings,
		"total_appointments": totalAppointments,
		"active_emergencies": activeEmergencies,
		"pending_approvals":  pendingApprovals,
		"open_tickets":       openTickets,
		"total_revenue":      totalRevenue,
	})
}

// ---- All Bookings (admin) ----

func AdminListBookings(c *gin.Context) {
	page, limit := parsePage(c)
	status := c.Query("status")
	ctx := context.Background()

	query := bson.M{}
	if status != "" {
		query["status"] = status
	}

	total, _ := database.Col(database.ColBookings).CountDocuments(ctx, query)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, _ := database.Col(database.ColBookings).Find(ctx, query, opts)
	var bookings []models.Booking
	cursor.All(ctx, &bookings)
	utils.PaginatedSuccessResponse(c, "Bookings fetched", bookings, page, limit, total)
}

func AdminUpdateBookingStatus(c *gin.Context) {
	bookingID := c.Param("bookingId")
	var req struct {
		Status         models.BookingStatus `json:"status"          binding:"required"`
		ProfessionalID string               `json:"professional_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	set := bson.M{"status": req.Status, "updated_at": time.Now()}
	if req.ProfessionalID != "" {
		set["professional_id"] = req.ProfessionalID
	}
	database.Col(database.ColBookings).UpdateOne(context.Background(),
		bson.M{"_id": bookingID}, bson.M{"$set": set})
	utils.SuccessResponse(c, "Booking updated", nil)
}
