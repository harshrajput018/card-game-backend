package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var rdb *redis.Client

func init() {
	// Connect to Redis server
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",                // No password set
		DB:       0,                 // Use default DB
	})
}

// LeaderboardEntry represents an entry in the leaderboard
type LeaderboardEntry struct {
	User   string
	Points int
	Wins   int // New field to store wins
}

// Handler for the leaderboard route
// Handler for the leaderboard route
func leaderboardHandler(w http.ResponseWriter, r *http.Request) {
    // Set CORS headers for the preflight request
    if r.Method == http.MethodOptions {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        w.WriteHeader(http.StatusOK)
        return
    }

    // Set CORS headers for the actual request
    w.Header().Set("Access-Control-Allow-Origin", "*")

    // Check if the request method is GET
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Retrieve leaderboard from Redis
    leaderboard, err := getLeaderboard()
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Display leaderboard
    fmt.Fprintf(w, "Leaderboard:\n")
    for _, entry := range leaderboard {
        fmt.Fprintf(w, "%s: Points: %d, Wins: %d\n", entry.User, entry.Points, entry.Wins)
    }
}

// IncrementLeaderboardPoints increments the points for a user in the leaderboard
func incrementLeaderboardPoints(user string) {
	err := rdb.ZIncrBy(ctx, "leaderboard", 1, user).Err()
	if err != nil {
		fmt.Println("Error incrementing leaderboard points:", err)
	}
}

// IncrementWins increments the wins for a user in the leaderboard
func incrementWins(user string) {
	err := rdb.HIncrBy(ctx, "wins", user, 1).Err()
	if err != nil {
		fmt.Println("Error incrementing wins:", err)
	}
}

// Retrieve leaderboard points and wins for all users
func getLeaderboard() ([]LeaderboardEntry, error) {

	
	// Retrieve leaderboard from Redis
	cmd := rdb.ZRevRangeWithScores(ctx, "leaderboard", 0, -1)
	result, err := cmd.Result()
	if err != nil {
		return nil, err
	}

	// Parse leaderboard entries
	var leaderboard []LeaderboardEntry
	for _, z := range result {
		score := z.Score
		points, _ := strconv.Atoi(fmt.Sprintf("%.0f", score))
		user := z.Member.(string)
		wins, _ := strconv.Atoi(rdb.HGet(ctx, "wins", user).Val()) // Retrieve wins from Redis
		leaderboard = append(leaderboard, LeaderboardEntry{
			User:   user,
			Points: points,
			Wins:   wins,
		})
	}

	fmt.Println(leaderboard)

	return leaderboard, nil
}

// StartGameHandler handles the request to start the game
func StartGameHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is OPTIONS (preflight request)
	if r.Method == http.MethodOptions {
		// Set appropriate CORS headers for preflight requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set CORS headers for the actual request
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Perform game setup logic (e.g., initialize game state)
	fmt.Println("Starting game for username:", req.Username)

	// Increment user's leaderboard points and wins
	incrementLeaderboardPoints(req.Username)
	incrementWins(req.Username) // Increment wins

	// Respond with success status
	w.WriteHeader(http.StatusOK)
}

func main() {
	// Seed initial leaderboard
	

	// Define routes
	http.HandleFunc("/leaderboard", leaderboardHandler)
	http.HandleFunc("/start", StartGameHandler)

	// Start the server
	fmt.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}

