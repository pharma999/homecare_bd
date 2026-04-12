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

// ── Subscription Plans ────────────────────────────────────────────────────

func AdminListSubscriptionPlans(c *gin.Context) {
	ctx := context.Background()
	cursor, err := database.Col(database.ColSubscriptionPlans).Find(ctx,
		bson.M{},
		options.Find().SetSort(bson.D{{Key: "display_order", Value: 1}}),
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch plans")
		return
	}
	var plans []models.SubscriptionPlan
	cursor.All(ctx, &plans)
	utils.SuccessResponse(c, "Subscription plans fetched", plans)
}

func AdminCreateSubscriptionPlan(c *gin.Context) {
	var req struct {
		Name             string   `json:"name"               binding:"required"`
		Description      string   `json:"description"`
		Duration         string   `json:"duration"           binding:"required"`
		Price            float64  `json:"price"              binding:"required"`
		Features         []string `json:"features"`
		MaxBookings      int      `json:"max_bookings"`
		MaxFamilyMembers int      `json:"max_family_members"`
		DisplayOrder     int      `json:"display_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	now := time.Now()
	plan := models.SubscriptionPlan{
		ID:               uuid.New().String(),
		Name:             req.Name,
		Description:      req.Description,
		Duration:         models.PlanDuration(req.Duration),
		Price:            req.Price,
		Features:         req.Features,
		MaxBookings:      req.MaxBookings,
		MaxFamilyMembers: req.MaxFamilyMembers,
		IsActive:         true,
		DisplayOrder:     req.DisplayOrder,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	database.Col(database.ColSubscriptionPlans).InsertOne(context.Background(), plan)
	utils.CreatedResponse(c, "Subscription plan created", plan)
}

func AdminUpdateSubscriptionPlan(c *gin.Context) {
	planID := c.Param("planId")
	var req struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		Price        float64  `json:"price"`
		Features     []string `json:"features"`
		IsActive     *bool    `json:"is_active"`
		DisplayOrder int      `json:"display_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	set := bson.M{"updated_at": time.Now()}
	if req.Name != "" {
		set["name"] = req.Name
	}
	if req.Description != "" {
		set["description"] = req.Description
	}
	if req.Price > 0 {
		set["price"] = req.Price
	}
	if req.Features != nil {
		set["features"] = req.Features
	}
	if req.IsActive != nil {
		set["is_active"] = *req.IsActive
	}
	if req.DisplayOrder > 0 {
		set["display_order"] = req.DisplayOrder
	}
	database.Col(database.ColSubscriptionPlans).UpdateOne(context.Background(),
		bson.M{"_id": planID}, bson.M{"$set": set})
	utils.SuccessResponse(c, "Plan updated", nil)
}

func AdminDeleteSubscriptionPlan(c *gin.Context) {
	planID := c.Param("planId")
	database.Col(database.ColSubscriptionPlans).DeleteOne(context.Background(),
		bson.M{"_id": planID})
	utils.SuccessResponse(c, "Plan deleted", nil)
}

// ── Support Tickets ───────────────────────────────────────────────────────

func AdminListSupportTickets(c *gin.Context) {
	page, limit := parsePage(c)
	status := c.Query("status")
	priority := c.Query("priority")
	category := c.Query("category")
	ctx := context.Background()

	query := bson.M{}
	if status != "" {
		query["status"] = status
	}
	if priority != "" {
		query["priority"] = priority
	}
	if category != "" {
		query["category"] = category
	}

	total, _ := database.Col(database.ColSupportTickets).CountDocuments(ctx, query)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, _ := database.Col(database.ColSupportTickets).Find(ctx, query, opts)
	var tickets []models.SupportTicket
	cursor.All(ctx, &tickets)
	utils.PaginatedSuccessResponse(c, "Tickets fetched", tickets, page, limit, total)
}

func AdminGetSupportTicket(c *gin.Context) {
	ticketID := c.Param("ticketId")
	var ticket models.SupportTicket
	err := database.Col(database.ColSupportTickets).FindOne(context.Background(),
		bson.M{"_id": ticketID}).Decode(&ticket)
	if err == mongo.ErrNoDocuments {
		utils.NotFoundResponse(c, "Ticket not found")
		return
	}
	utils.SuccessResponse(c, "Ticket fetched", ticket)
}

func AdminUpdateSupportTicket(c *gin.Context) {
	ticketID := c.Param("ticketId")
	var req struct {
		Status     string `json:"status"`
		Priority   string `json:"priority"`
		AssignedTo string `json:"assigned_to"`
		Resolution string `json:"resolution"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	set := bson.M{"updated_at": time.Now()}
	if req.Status != "" {
		set["status"] = req.Status
		if req.Status == string(models.TicketResolved) || req.Status == string(models.TicketClosed) {
			now := time.Now()
			set["resolved_at"] = now
		}
	}
	if req.Priority != "" {
		set["priority"] = req.Priority
	}
	if req.AssignedTo != "" {
		set["assigned_to"] = req.AssignedTo
	}
	if req.Resolution != "" {
		set["resolution"] = req.Resolution
	}

	database.Col(database.ColSupportTickets).UpdateOne(context.Background(),
		bson.M{"_id": ticketID}, bson.M{"$set": set})
	utils.SuccessResponse(c, "Ticket updated", nil)
}

// User submits a support ticket
func CreateSupportTicket(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		Subject     string `json:"subject"     binding:"required"`
		Description string `json:"description" binding:"required"`
		Category    string `json:"category"    binding:"required"`
		ReferenceID string `json:"reference_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	var user models.User
	database.Col(database.ColUsers).FindOne(context.Background(),
		bson.M{"_id": userID}).Decode(&user)

	now := time.Now()
	ticket := models.SupportTicket{
		ID:          uuid.New().String(),
		UserID:      userID,
		UserName:    user.Name,
		UserPhone:   user.PhoneNumber,
		Subject:     req.Subject,
		Description: req.Description,
		Category:    req.Category,
		Status:      models.TicketOpen,
		Priority:    models.TicketMedium,
		ReferenceID: req.ReferenceID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	database.Col(database.ColSupportTickets).InsertOne(context.Background(), ticket)
	utils.CreatedResponse(c, "Support ticket created", ticket)
}

func GetMyTickets(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, limit := parsePage(c)
	ctx := context.Background()

	query := bson.M{"user_id": userID}
	total, _ := database.Col(database.ColSupportTickets).CountDocuments(ctx, query)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, _ := database.Col(database.ColSupportTickets).Find(ctx, query, opts)
	var tickets []models.SupportTicket
	cursor.All(ctx, &tickets)
	utils.PaginatedSuccessResponse(c, "Tickets fetched", tickets, page, limit, total)
}

// ── Platform Settings (Super Admin only) ─────────────────────────────────

func GetPlatformSettings(c *gin.Context) {
	ctx := context.Background()
	var settings models.PlatformSettings
	err := database.Col(database.ColPlatformSettings).FindOne(ctx,
		bson.M{}).Decode(&settings)
	if err == mongo.ErrNoDocuments {
		// return defaults
		settings = models.PlatformSettings{
			DoctorCommissionPct:  15.0,
			NurseCommissionPct:   12.0,
			BookingCommissionPct: 10.0,
			EmergencyBaseFee:     500.0,
			SupportEmail:         "support@homecare.com",
			SupportPhone:         "+91-1800-000-0000",
			AppVersion:           "1.0.0",
			MaintenanceMode:      false,
		}
	}
	utils.SuccessResponse(c, "Settings fetched", settings)
}

func UpdatePlatformSettings(c *gin.Context) {
	adminID := middleware.GetUserID(c)
	var req struct {
		DoctorCommissionPct  *float64 `json:"doctor_commission_pct"`
		NurseCommissionPct   *float64 `json:"nurse_commission_pct"`
		BookingCommissionPct *float64 `json:"booking_commission_pct"`
		EmergencyBaseFee     *float64 `json:"emergency_base_fee"`
		SupportEmail         string   `json:"support_email"`
		SupportPhone         string   `json:"support_phone"`
		AppVersion           string   `json:"app_version"`
		MaintenanceMode      *bool    `json:"maintenance_mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	set := bson.M{"updated_by": adminID, "updated_at": time.Now()}
	if req.DoctorCommissionPct != nil {
		set["doctor_commission_pct"] = *req.DoctorCommissionPct
	}
	if req.NurseCommissionPct != nil {
		set["nurse_commission_pct"] = *req.NurseCommissionPct
	}
	if req.BookingCommissionPct != nil {
		set["booking_commission_pct"] = *req.BookingCommissionPct
	}
	if req.EmergencyBaseFee != nil {
		set["emergency_base_fee"] = *req.EmergencyBaseFee
	}
	if req.SupportEmail != "" {
		set["support_email"] = req.SupportEmail
	}
	if req.SupportPhone != "" {
		set["support_phone"] = req.SupportPhone
	}
	if req.AppVersion != "" {
		set["app_version"] = req.AppVersion
	}
	if req.MaintenanceMode != nil {
		set["maintenance_mode"] = *req.MaintenanceMode
	}

	ctx := context.Background()
	result, _ := database.Col(database.ColPlatformSettings).UpdateOne(ctx,
		bson.M{}, bson.M{"$set": set})
	if result.MatchedCount == 0 {
		// First time — insert
		database.Col(database.ColPlatformSettings).InsertOne(ctx, bson.M{
			"_id":                   uuid.New().String(),
			"doctor_commission_pct":  15.0,
			"nurse_commission_pct":   12.0,
			"booking_commission_pct": 10.0,
			"emergency_base_fee":     500.0,
			"support_email":          "support@homecare.com",
			"support_phone":          "+91-1800-000-0000",
			"app_version":            "1.0.0",
			"maintenance_mode":       false,
			"updated_by":             adminID,
			"updated_at":             time.Now(),
		})
	}
	utils.SuccessResponse(c, "Settings updated", nil)
}

// ── Service Zones ─────────────────────────────────────────────────────────

func AdminListZones(c *gin.Context) {
	ctx := context.Background()
	status := c.Query("status")
	query := bson.M{}
	if status != "" {
		query["status"] = status
	}
	cursor, _ := database.Col(database.ColServiceZones).Find(ctx, query,
		options.Find().SetSort(bson.D{{Key: "city", Value: 1}}))
	var zones []models.ServiceZone
	cursor.All(ctx, &zones)
	utils.SuccessResponse(c, "Zones fetched", zones)
}

func AdminCreateZone(c *gin.Context) {
	adminID := middleware.GetUserID(c)
	var req struct {
		Name     string   `json:"name"      binding:"required"`
		City     string   `json:"city"      binding:"required"`
		State    string   `json:"state"     binding:"required"`
		PinCodes []string `json:"pin_codes"`
		Status   string   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	now := time.Now()
	status := models.ZonePlanned
	if req.Status != "" {
		status = models.ZoneStatus(req.Status)
	}
	zone := models.ServiceZone{
		ID:        uuid.New().String(),
		Name:      req.Name,
		City:      req.City,
		State:     req.State,
		PinCodes:  req.PinCodes,
		Status:    status,
		CreatedBy: adminID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	database.Col(database.ColServiceZones).InsertOne(context.Background(), zone)
	utils.CreatedResponse(c, "Zone created", zone)
}

func AdminUpdateZone(c *gin.Context) {
	zoneID := c.Param("zoneId")
	var req struct {
		Name     string   `json:"name"`
		PinCodes []string `json:"pin_codes"`
		Status   string   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	set := bson.M{"updated_at": time.Now()}
	if req.Name != "" {
		set["name"] = req.Name
	}
	if req.PinCodes != nil {
		set["pin_codes"] = req.PinCodes
	}
	if req.Status != "" {
		set["status"] = req.Status
	}
	database.Col(database.ColServiceZones).UpdateOne(context.Background(),
		bson.M{"_id": zoneID}, bson.M{"$set": set})
	utils.SuccessResponse(c, "Zone updated", nil)
}

// ── Admin User Management (Super Admin only) ──────────────────────────────

func SuperAdminListAdmins(c *gin.Context) {
	ctx := context.Background()
	page, limit := parsePage(c)
	query := bson.M{"role": bson.M{"$in": []string{
		string(models.RoleAdmin), string(models.RoleSuperAdmin),
	}}}
	total, _ := database.Col(database.ColUsers).CountDocuments(ctx, query)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))
	cursor, _ := database.Col(database.ColUsers).Find(ctx, query, opts)
	var users []models.User
	cursor.All(ctx, &users)
	responses := make([]models.UserResponse, 0, len(users))
	for i := range users {
		responses = append(responses, buildUserResponse(&users[i]))
	}
	utils.PaginatedSuccessResponse(c, "Admins fetched", responses, page, limit, total)
}

func SuperAdminCreateAdmin(c *gin.Context) {
	var req models.AdminCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	if req.Role != string(models.RoleAdmin) && req.Role != string(models.RoleSuperAdmin) {
		utils.BadRequestResponse(c, "role must be ADMIN or SUPER_ADMIN")
		return
	}
	now := time.Now()
	user := models.User{
		ID:            uuid.New().String(),
		Name:          req.Name,
		PhoneNumber:   req.PhoneNumber,
		Email:         req.Email,
		Role:          models.UserRole(req.Role),
		Status:        models.UserStatusActive,
		BlockStatus:   models.BlockStatusUnblocked,
		UserService:   models.UserServiceUnsubscribed,
		ServiceStatus: models.ServiceStatusNew,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	_, err := database.Col(database.ColUsers).InsertOne(context.Background(), user)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create admin (phone may be duplicate)")
		return
	}
	utils.CreatedResponse(c, "Admin user created", buildUserResponse(&user))
}

func SuperAdminDeleteAdmin(c *gin.Context) {
	adminID := c.Param("adminId")
	// Prevent self-deletion
	callerID := middleware.GetUserID(c)
	if adminID == callerID {
		utils.BadRequestResponse(c, "Cannot delete yourself")
		return
	}
	database.Col(database.ColUsers).DeleteOne(context.Background(),
		bson.M{"_id": adminID, "role": bson.M{"$in": []string{
			string(models.RoleAdmin), string(models.RoleSuperAdmin),
		}}})
	utils.SuccessResponse(c, "Admin removed", nil)
}

// ── Extended Analytics (Super Admin) ─────────────────────────────────────

func SuperAdminRevenueReport(c *gin.Context) {
	ctx := context.Background()

	// Monthly revenue aggregation
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"status": "COMPLETED"}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"year":  bson.M{"$year": "$created_at"},
				"month": bson.M{"$month": "$created_at"},
			},
			"revenue":       bson.M{"$sum": "$amount"},
			"count":         bson.M{"$sum": 1},
			"avg_order":     bson.M{"$avg": "$amount"},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "_id.year", Value: -1}, {Key: "_id.month", Value: -1}}}},
		{{Key: "$limit", Value: 12}},
	}
	cursor, _ := database.Col(database.ColPayments).Aggregate(ctx, pipeline)
	var monthly []bson.M
	cursor.All(ctx, &monthly)

	// Total stats
	totalRevPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"status": "COMPLETED"}}},
		{{Key: "$group", Value: bson.M{
			"_id":        nil,
			"total":      bson.M{"$sum": "$amount"},
			"count":      bson.M{"$sum": 1},
			"avg_order":  bson.M{"$avg": "$amount"},
		}}},
	}
	cursor2, _ := database.Col(database.ColPayments).Aggregate(ctx, totalRevPipeline)
	var totals []bson.M
	cursor2.All(ctx, &totals)

	totalRevenue := 0.0
	totalOrders := int64(0)
	if len(totals) > 0 {
		if v, ok := totals[0]["total"].(float64); ok {
			totalRevenue = v
		}
		if v, ok := totals[0]["count"].(int32); ok {
			totalOrders = int64(v)
		}
	}

	// User growth
	userCount, _ := database.Col(database.ColUsers).CountDocuments(ctx, bson.M{})
	activeSubCount, _ := database.Col(database.ColUserSubscriptions).CountDocuments(ctx,
		bson.M{"status": models.SubStatusActive})

	utils.SuccessResponse(c, "Revenue report", gin.H{
		"total_revenue":       totalRevenue,
		"total_transactions":  totalOrders,
		"total_users":         userCount,
		"active_subscriptions": activeSubCount,
		"monthly_breakdown":   monthly,
	})
}

// ── User Subscriptions ────────────────────────────────────────────────────

func AdminListUserSubscriptions(c *gin.Context) {
	page, limit := parsePage(c)
	status := c.Query("status")
	ctx := context.Background()

	query := bson.M{}
	if status != "" {
		query["status"] = status
	}

	total, _ := database.Col(database.ColUserSubscriptions).CountDocuments(ctx, query)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, _ := database.Col(database.ColUserSubscriptions).Find(ctx, query, opts)
	var subs []models.UserSubscription
	cursor.All(ctx, &subs)
	utils.PaginatedSuccessResponse(c, "Subscriptions fetched", subs, page, limit, total)
}

// ── Appointments Admin ────────────────────────────────────────────────────

func AdminListAppointments(c *gin.Context) {
	page, limit := parsePage(c)
	status := c.Query("status")
	ctx := context.Background()

	query := bson.M{}
	if status != "" {
		query["status"] = status
	}

	total, _ := database.Col(database.ColAppointments).CountDocuments(ctx, query)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, _ := database.Col(database.ColAppointments).Find(ctx, query, opts)
	var appointments []models.Appointment
	cursor.All(ctx, &appointments)
	utils.PaginatedSuccessResponse(c, "Appointments fetched", appointments, page, limit, total)
}
