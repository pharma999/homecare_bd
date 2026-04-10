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

func CreateBooking(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.CreateBookingRequest
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

	var scheduledAt *time.Time
	if req.ScheduledAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err == nil {
			scheduledAt = &t
		}
	}

	now := time.Now()
	booking := models.Booking{
		ID:               uuid.New().String(),
		PatientUserID:    userID,
		ProfessionalID:   req.ProfessionalID,
		ServiceID:        req.ServiceID,
		BookingType:      req.BookingType,
		Status:           models.BookingStatusPending,
		ScheduledAt:      scheduledAt,
		PatientAddress:   req.PatientAddress,
		PatientLatitude:  req.PatientLatitude,
		PatientLongitude: req.PatientLongitude,
		TotalAmount:      svc.BasePrice,
		PaymentStatus:    "PENDING",
		Notes:            req.Notes,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	database.Col(database.ColBookings).InsertOne(ctx, booking)
	createNotification(ctx, userID, "Booking Created",
		"Your booking has been placed. A professional will be assigned shortly.", "BOOKING", booking.ID)
	utils.CreatedResponse(c, "Booking created", booking)
}

func GetMyBookings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, limit := parsePage(c)
	statusFilter := c.Query("status")
	ctx := context.Background()

	query := bson.M{"patient_user_id": userID}
	if statusFilter != "" {
		query["status"] = statusFilter
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

func GetBooking(c *gin.Context) {
	bookingID := c.Param("bookingId")
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	ctx := context.Background()

	var booking models.Booking
	if err := database.Col(database.ColBookings).FindOne(ctx, bson.M{"_id": bookingID}).Decode(&booking); err != nil {
		utils.NotFoundResponse(c, "Booking not found")
		return
	}
	isAdmin := role == string(models.RoleAdmin) || role == string(models.RoleSuperAdmin)
	if booking.PatientUserID != userID && !isAdmin {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}
	utils.SuccessResponse(c, "Booking fetched", booking)
}

func CancelBooking(c *gin.Context) {
	bookingID := c.Param("bookingId")
	userID := middleware.GetUserID(c)
	ctx := context.Background()

	var booking models.Booking
	if err := database.Col(database.ColBookings).FindOne(ctx,
		bson.M{"_id": bookingID, "patient_user_id": userID}).Decode(&booking); err != nil {
		utils.NotFoundResponse(c, "Booking not found")
		return
	}
	if booking.Status == models.BookingStatusCompleted || booking.Status == models.BookingStatusCancelled {
		utils.BadRequestResponse(c, "Booking cannot be cancelled in current state")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)
	database.Col(database.ColBookings).UpdateOne(ctx, bson.M{"_id": bookingID},
		bson.M{"$set": bson.M{"status": models.BookingStatusCancelled,
			"cancelled_reason": req.Reason, "updated_at": time.Now()}})
	createNotification(ctx, userID, "Booking Cancelled", "Your booking has been cancelled.", "BOOKING", bookingID)
	utils.SuccessResponse(c, "Booking cancelled", nil)
}

// ---- Cart ----

func GetCart(c *gin.Context) {
	userID := middleware.GetUserID(c)
	cursor, _ := database.Col(database.ColCartItems).Find(context.Background(), bson.M{"user_id": userID})
	var items []models.CartItem
	cursor.All(context.Background(), &items)

	total := 0.0
	for _, it := range items {
		total += it.Price * float64(it.Quantity)
	}
	utils.SuccessResponse(c, "Cart fetched", gin.H{"items": items, "total": total, "count": len(items)})
}

func AddToCart(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	qty := req.Quantity
	if qty < 1 {
		qty = 1
	}
	ctx := context.Background()

	var existing models.CartItem
	err := database.Col(database.ColCartItems).FindOne(ctx,
		bson.M{"user_id": userID, "service_id": req.ServiceID}).Decode(&existing)

	if err == mongo.ErrNoDocuments {
		item := models.CartItem{
			ID:        uuid.New().String(),
			UserID:    userID,
			ServiceID: req.ServiceID,
			Title:     req.Title,
			Price:     req.Price,
			Quantity:  qty,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		database.Col(database.ColCartItems).InsertOne(ctx, item)
		utils.CreatedResponse(c, "Item added to cart", item)
	} else if err == nil {
		database.Col(database.ColCartItems).UpdateOne(ctx,
			bson.M{"_id": existing.ID},
			bson.M{"$set": bson.M{"quantity": existing.Quantity + qty, "updated_at": time.Now()}})
		utils.SuccessResponse(c, "Cart updated", nil)
	} else {
		utils.InternalServerErrorResponse(c, "Cart operation failed")
	}
}

func UpdateCartQuantity(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.UpdateCartQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	ctx := context.Background()
	if req.Quantity < 1 {
		database.Col(database.ColCartItems).DeleteOne(ctx,
			bson.M{"user_id": userID, "service_id": req.ServiceID})
		utils.SuccessResponse(c, "Item removed from cart", nil)
		return
	}
	database.Col(database.ColCartItems).UpdateOne(ctx,
		bson.M{"user_id": userID, "service_id": req.ServiceID},
		bson.M{"$set": bson.M{"quantity": req.Quantity, "updated_at": time.Now()}})
	utils.SuccessResponse(c, "Quantity updated", nil)
}

func RemoveFromCart(c *gin.Context) {
	userID := middleware.GetUserID(c)
	database.Col(database.ColCartItems).DeleteOne(context.Background(),
		bson.M{"user_id": userID, "service_id": c.Param("serviceId")})
	utils.SuccessResponse(c, "Item removed", nil)
}

func ClearCart(c *gin.Context) {
	userID := middleware.GetUserID(c)
	database.Col(database.ColCartItems).DeleteMany(context.Background(), bson.M{"user_id": userID})
	utils.SuccessResponse(c, "Cart cleared", nil)
}

func CheckoutCart(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		PatientAddress   string `json:"patient_address"   binding:"required"`
		PatientLatitude  string `json:"patient_latitude"`
		PatientLongitude string `json:"patient_longitude"`
		BookingType      string `json:"booking_type"`
		ScheduledAt      string `json:"scheduled_at"`
		Notes            string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	ctx := context.Background()

	cursor, _ := database.Col(database.ColCartItems).Find(ctx, bson.M{"user_id": userID})
	var items []models.CartItem
	cursor.All(ctx, &items)
	if len(items) == 0 {
		utils.BadRequestResponse(c, "Cart is empty")
		return
	}

	bookingType := req.BookingType
	if bookingType == "" {
		bookingType = "QUICK"
	}
	var scheduledAt *time.Time
	if req.ScheduledAt != "" {
		t, err := time.Parse(time.RFC3339, req.ScheduledAt)
		if err == nil {
			scheduledAt = &t
		}
	}

	var created []models.Booking
	now := time.Now()
	for _, item := range items {
		b := models.Booking{
			ID:               uuid.New().String(),
			PatientUserID:    userID,
			ServiceID:        item.ServiceID,
			BookingType:      bookingType,
			Status:           models.BookingStatusPending,
			ScheduledAt:      scheduledAt,
			PatientAddress:   req.PatientAddress,
			PatientLatitude:  req.PatientLatitude,
			PatientLongitude: req.PatientLongitude,
			TotalAmount:      item.Price * float64(item.Quantity),
			PaymentStatus:    "PENDING",
			Notes:            req.Notes,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		database.Col(database.ColBookings).InsertOne(ctx, b)
		created = append(created, b)
	}
	database.Col(database.ColCartItems).DeleteMany(ctx, bson.M{"user_id": userID})
	createNotification(ctx, userID, "Order Placed",
		"Your services have been booked. Professionals will be assigned shortly.", "BOOKING", "")
	utils.CreatedResponse(c, "Checkout successful", gin.H{"bookings": created})
}

// ---- Payments ----

func InitiatePayment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req struct {
		BookingID     string  `json:"booking_id"`
		AppointmentID string  `json:"appointment_id"`
		Amount        float64 `json:"amount"         binding:"required"`
		PaymentMethod string  `json:"payment_method" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	now := time.Now()
	payment := models.Payment{
		ID:            uuid.New().String(),
		UserID:        userID,
		BookingID:     req.BookingID,
		AppointmentID: req.AppointmentID,
		Amount:        req.Amount,
		Currency:      "INR",
		PaymentMethod: req.PaymentMethod,
		Status:        "PENDING",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	database.Col(database.ColPayments).InsertOne(context.Background(), payment)
	utils.CreatedResponse(c, "Payment initiated", gin.H{
		"payment_id":     payment.ID,
		"amount":         payment.Amount,
		"status":         "PENDING",
		"payment_method": payment.PaymentMethod,
	})
}

func ConfirmPayment(c *gin.Context) {
	paymentID := c.Param("paymentId")
	var req struct {
		TransactionID string `json:"transaction_id" binding:"required"`
		Status        string `json:"status"         binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}
	ctx := context.Background()
	now := time.Now()
	database.Col(database.ColPayments).UpdateOne(ctx, bson.M{"_id": paymentID},
		bson.M{"$set": bson.M{
			"transaction_id": req.TransactionID,
			"status":         req.Status,
			"paid_at":        now,
			"updated_at":     now,
		}})

	var payment models.Payment
	database.Col(database.ColPayments).FindOne(ctx, bson.M{"_id": paymentID}).Decode(&payment)
	if payment.BookingID != "" {
		database.Col(database.ColBookings).UpdateOne(ctx, bson.M{"_id": payment.BookingID},
			bson.M{"$set": bson.M{"payment_status": req.Status}})
	}
	if payment.AppointmentID != "" {
		database.Col(database.ColAppointments).UpdateOne(ctx, bson.M{"_id": payment.AppointmentID},
			bson.M{"$set": bson.M{"payment_status": req.Status}})
	}
	utils.SuccessResponse(c, "Payment updated", nil)
}

func GetPaymentHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, _ := database.Col(database.ColPayments).Find(context.Background(), bson.M{"user_id": userID}, opts)
	var payments []models.Payment
	cursor.All(context.Background(), &payments)
	utils.SuccessResponse(c, "Payment history fetched", payments)
}

// ---- shared helpers ----

func createNotification(ctx context.Context, userID, title, body, notifType, refID string) {
	n := models.Notification{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       title,
		Body:        body,
		Type:        notifType,
		ReferenceID: refID,
		IsRead:      false,
		CreatedAt:   time.Now(),
	}
	database.Col(database.ColNotifications).InsertOne(ctx, n)
}

func parsePage(c *gin.Context) (int, int) {
	page := 1
	limit := 20
	if v := c.Query("page"); v != "" {
		if n, err := parseInt(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("limit"); v != "" {
		if n, err := parseInt(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	return page, limit
}

func parseInt(s string) (int, error) {
	var n int
	_, err := scanInt(s, &n)
	return n, err
}

func scanInt(s string, n *int) (int, error) {
	*n = 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errNotInt
		}
		*n = *n*10 + int(c-'0')
	}
	return *n, nil
}

var errNotInt = &parseError{}

type parseError struct{}

func (e *parseError) Error() string { return "not an integer" }
