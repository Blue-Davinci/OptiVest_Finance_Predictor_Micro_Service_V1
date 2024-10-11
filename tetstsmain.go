package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

type App struct {
	RedisDB           *redis.Client
	WebSocketUpgrader websocket.Upgrader
	Clients           map[string]*websocket.Conn
	Mutex             sync.Mutex
}

// Initialize Redis
func NewApp() *App {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	return &App{
		RedisDB: redisClient,
		Clients: make(map[string]*websocket.Conn),
	}
}

// WebSocket Handler
func (app *App) wsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	conn, err := app.WebSocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Register WebSocket connection
	app.Mutex.Lock()
	app.Clients[userID] = conn
	app.Mutex.Unlock()

	// Preload and send pending notifications
	go app.sendPendingNotifications(userID, conn)

	// Listen for WebSocket and Pub/Sub messages
	go app.listenForMessages(conn)
	go app.listenForPubSubMessages(userID, conn)
}

// Listen for WebSocket messages (e.g., clearing notifications)
func (app *App) listenForMessages(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket error: %v", err)
			conn.Close()
			return
		}

		// Handle client-sent messages (e.g., marking notifications as read)
		fmt.Println("Received from client:", string(message))
		app.handleClientMessage(string(message))
	}
}

// Pub/Sub: Listen for live messages
func (app *App) listenForPubSubMessages(userID string, conn *websocket.Conn) {
	pubSub := app.RedisDB.Subscribe(context.Background(), fmt.Sprintf("user:%s:notifications", userID))
	defer pubSub.Close()

	for msg := range pubSub.Channel() {
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
		if err != nil {
			log.Printf("WebSocket send error: %v", err)
			conn.Close()
			return
		}
	}
}

// Send pending notifications (when user logs in)
func (app *App) sendPendingNotifications(userID string, conn *websocket.Conn) {
	ctx := context.Background()

	// Check Redis for pending notifications
	pending, err := app.RedisDB.LRange(ctx, fmt.Sprintf("user:%s:pending_notifications", userID), 0, -1).Result()
	if err != nil {
		log.Printf("Redis error: %v", err)
		return
	}

	// Send each pending notification via WebSocket
	for _, notification := range pending {
		err := conn.WriteMessage(websocket.TextMessage, []byte(notification))
		if err != nil {
			log.Printf("WebSocket send error: %v", err)
			conn.Close()
			return
		}
	}

	// After sending, clear pending notifications from Redis
	app.RedisDB.Del(ctx, fmt.Sprintf("user:%s:pending_notifications", userID))
}

// Handle client-sent messages (e.g., notification read)
func (app *App) handleClientMessage(message string) {
	// For example, this method could handle marking notifications as read
	// Assuming the message contains a notification ID
	fmt.Println("Client read notification:", message)

	// Implement logic to mark notification as read in the database (PostgreSQL)
	app.markNotificationAsReadInDB(message)
}

// Notify users (for group invites or market updates)
func (app *App) sendNotification(userID, message string) {
	ctx := context.Background()

	// If the user is online, send via WebSocket
	app.Mutex.Lock()
	conn, online := app.Clients[userID]
	app.Mutex.Unlock()

	if online {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Printf("WebSocket send error: %v", err)
		}
	} else {
		// If offline, save to Redis for future delivery (pending notifications)
		app.RedisDB.RPush(ctx, fmt.Sprintf("user:%s:pending_notifications", userID), message)
	}

	// Save to PostgreSQL for persistence
	app.saveNotificationToDB(userID, message)
}

// Mark notifications as read in PostgreSQL
func (app *App) markNotificationAsReadInDB(notificationID string) {
	// This is a placeholder for saving to PostgreSQL
	fmt.Println("Marking notification as read in PostgreSQL:", notificationID)
	// Actual database logic would go here
}

// Save notifications to PostgreSQL
func (app *App) saveNotificationToDB(userID, message string) {
	// This is a placeholder for saving to PostgreSQL
	fmt.Printf("Saving notification for user %s to PostgreSQL: %s\n", userID, message)
	// Actual database logic would go here
}

// Preload followed stocks for users at login
func (app *App) preloadUserFollows(userID string, stocks []string) {
	ctx := context.Background()
	for _, symbol := range stocks {
		app.RedisDB.SAdd(ctx, "stock:"+symbol+":followers", userID)
	}
}

// Simulate Market Data Updates
func (app *App) simulateMarketDataUpdates() {
	for {
		time.Sleep(10 * time.Second) // Simulate market data update

		// Simulate market data for AAPL and MSFT
		app.notifyMarketFollowers("AAPL", "New AAPL news!")
		app.notifyMarketFollowers("MSFT", "New MSFT news!")
	}
}

// Notify users about market updates
func (app *App) notifyMarketFollowers(stockID string, news string) {
	ctx := context.Background()

	// Retrieve followers from Redis
	followers, err := app.RedisDB.SMembers(ctx, "stock:"+stockID+":followers").Result()
	if err != nil {
		log.Printf("Redis error: %v", err)
		return
	}

	// Notify each follower
	for _, userID := range followers {
		app.sendNotification(userID, fmt.Sprintf("Market Update: %s", news))
	}
}

func main() {
	app := NewApp()

	// Preload followed stocks for two users
	app.preloadUserFollows("user1", []string{"AAPL"})
	app.preloadUserFollows("user2", []string{"MSFT"})

	// Simulate market data updates
	go app.simulateMarketDataUpdates()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Simulate user ID middleware
		userID := r.URL.Query().Get("userID") // Get userID from query params
		ctx := context.WithValue(r.Context(), "userID", userID)
		app.wsHandler(w, r.WithContext(ctx))
	})

	log.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", nil)
}
