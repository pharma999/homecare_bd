package handlers

import (
	"context"
	"strings"
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

// ChatMessage stores a single turn in a conversation.
type ChatMessage struct {
	ID        string    `bson:"_id,omitempty" json:"message_id"`
	UserID    string    `bson:"user_id"       json:"user_id"`
	SessionID string    `bson:"session_id"    json:"session_id"`
	Role      string    `bson:"role"          json:"role"` // "user" | "bot"
	Content   string    `bson:"content"       json:"content"`
	CreatedAt time.Time `bson:"created_at"    json:"created_at"`
}

const colChat = "chat_messages"

// SendChatMessage handles POST /api/chatbot/message
// It matches intent from the user's message and returns a helpful reply.
func SendChatMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Message   string `json:"message"    binding:"required"`
		SessionID string `json:"session_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	ctx := context.Background()

	// Persist user message
	userMsg := ChatMessage{
		ID:        uuid.New().String(),
		UserID:    userID,
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Message,
		CreatedAt: time.Now(),
	}
	database.Col(colChat).InsertOne(ctx, userMsg)

	// Generate rule-based response enriched with live DB data
	reply := buildBotReply(ctx, userID, req.Message)

	// Persist bot reply
	botMsg := ChatMessage{
		ID:        uuid.New().String(),
		UserID:    userID,
		SessionID: sessionID,
		Role:      "bot",
		Content:   reply,
		CreatedAt: time.Now(),
	}
	database.Col(colChat).InsertOne(ctx, botMsg)

	utils.SuccessResponse(c, "Message sent", gin.H{
		"session_id": sessionID,
		"message_id": botMsg.ID,
		"reply":      reply,
	})
}

// GetChatHistory handles GET /api/chatbot/history?session_id=<id>
func GetChatHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	sessionID := c.Query("session_id")

	query := bson.M{"user_id": userID}
	if sessionID != "" {
		query["session_id"] = sessionID
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetLimit(100)

	cursor, _ := database.Col(colChat).Find(context.Background(), query, opts)
	var msgs []ChatMessage
	cursor.All(context.Background(), &msgs)
	utils.SuccessResponse(c, "Chat history fetched", msgs)
}

// buildBotReply returns a context-aware response based on simple intent matching.
func buildBotReply(ctx context.Context, userID, msg string) string {
	lower := strings.ToLower(msg)

	switch {
	// Booking intent
	case containsAny(lower, "book", "booking", "schedule", "appointment"):
		return "I can help you book a service or appointment. " +
			"Please go to the Services section to choose a service, or tap Appointments to schedule with a doctor."

	// Emergency / SOS
	case containsAny(lower, "emergency", "sos", "ambulance", "urgent", "help"):
		return "If this is a medical emergency, please use the SOS button on the home screen immediately — " +
			"it will alert the nearest available ambulance and notify your family. " +
			"For non-urgent queries, I'm here to help."

	// Booking status
	case containsAny(lower, "status", "my booking", "my order", "track"):
		count, _ := database.Col(database.ColBookings).CountDocuments(ctx,
			bson.M{"patient_user_id": userID, "status": bson.M{"$in": []string{"PENDING", "ACCEPTED", "IN_PROGRESS"}}})
		if count > 0 {
			return "You have active bookings. Head to Profile → Bookings to view the latest status and track your professional."
		}
		return "You don't have any active bookings right now. Would you like to book a new service?"

	// Payment
	case containsAny(lower, "payment", "pay", "invoice", "bill", "charge"):
		return "You can view your payment history under Profile → Payments. " +
			"If you have a payment issue, please raise a support ticket and our team will respond within 24 hours."

	// Doctors / professionals
	case containsAny(lower, "doctor", "physician", "specialist", "nurse", "caregiver"):
		return "You can search for doctors and nurses from the Appointments section. " +
			"Filter by specialty, consultation type, fees, and rating to find the right match."

	// Services
	case containsAny(lower, "service", "what do you offer", "available service"):
		cats := fetchCategoryNames(ctx)
		if len(cats) > 0 {
			return "We currently offer: " + strings.Join(cats, ", ") + ". " +
				"Tap Services on the home screen to explore and book."
		}
		return "We offer a range of home-care services including nursing, doctor visits, lab tests, and more. " +
			"Tap Services on the home screen to explore."

	// Medical records
	case containsAny(lower, "record", "report", "prescription", "lab", "test result"):
		return "Your medical records and prescriptions are available under Profile → Records. " +
			"You can upload new reports and share them with your care team."

	// Support / complaint
	case containsAny(lower, "support", "complaint", "issue", "problem", "help me"):
		return "I'm sorry to hear you're facing an issue. " +
			"Please raise a support ticket from Profile → Support and our team will assist you within 24 hours."

	// Profile
	case containsAny(lower, "profile", "account", "edit", "update", "address"):
		return "You can update your profile, address, and personal details from the Profile section. " +
			"Tap the profile icon at the bottom navigation bar."

	// Greeting
	case containsAny(lower, "hello", "hi", "hey", "good morning", "good evening"):
		return "Hello! I'm your HomeCare assistant. I can help you with bookings, appointments, " +
			"emergencies, medical records, payments, and more. How can I help you today?"

	// Farewell
	case containsAny(lower, "bye", "goodbye", "thank", "thanks"):
		return "You're welcome! Stay safe and healthy. Feel free to reach out any time. 😊"

	default:
		return "I'm not sure I understood that. I can help you with:\n" +
			"• Booking services or appointments\n" +
			"• Emergency SOS assistance\n" +
			"• Tracking booking status\n" +
			"• Medical records\n" +
			"• Payments\n" +
			"• Support tickets\n\n" +
			"What would you like to do?"
	}
}

func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

func fetchCategoryNames(ctx context.Context) []string {
	cursor, err := database.Col(database.ColCategories).Find(ctx,
		bson.M{"is_active": true},
		options.Find().SetProjection(bson.M{"name": 1}).SetLimit(8),
	)
	if err != nil {
		return nil
	}
	var cats []models.ServiceCategory
	cursor.All(ctx, &cats)
	names := make([]string, 0, len(cats))
	for _, cat := range cats {
		names = append(names, cat.Name)
	}
	return names
}
