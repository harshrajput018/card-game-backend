package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

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
}

// Handler for the leaderboard route
func leaderboardHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve leaderboard from Redis
	leaderboard, err := getLeaderboard()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Display leaderboard
	fmt.Fprintf(w, "Leaderboard:\n")
	for _, entry := range leaderboard {
		fmt.Fprintf(w, "%s: %d\n", entry.User, entry.Points)
	}
}

// IncrementLeaderboardPoints increments the points for a user in the leaderboard
func incrementLeaderboardPoints(user string) {
	err := rdb.ZIncrBy(ctx, "leaderboard", 1, user).Err()
	if err != nil {
		fmt.Println("Error incrementing leaderboard points:", err)
	}
}

// Retrieve leaderboard points for all users
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
		leaderboard = append(leaderboard, LeaderboardEntry{
			User:   z.Member.(string),
			Points: points,
		})
	}

	fmt.Println(leaderboard)

	return leaderboard, nil
}

func main() {
	// Seed initial leaderboard
	seedLeaderboard()

	// Define routes
	http.HandleFunc("/leaderboard", leaderboardHandler)

	// Start the server
	fmt.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}

// SeedLeaderboard initializes the leaderboard with sample data
func seedLeaderboard() {
	// Simulate some activity to seed the leaderboard
	users := []string{"user1", "user2", "user3", "user4", "user5"}
	for _, user := range users {
		for i := 0; i < 10; i++ {
			incrementLeaderboardPoints(user)
			time.Sleep(time.Millisecond * 100) // Simulate activity delay
		}
	}
}
