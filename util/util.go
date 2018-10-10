package util

import (
	"net/http"
	"time"
	"math"
	"encoding/json"
	"fmt"
	"math/rand"
)

func Message(status bool, message string) (map[string]interface{}) {
	return map[string]interface{} {"status" : status, "message" : message}
}

func InvalidRequestMessage() map[string] interface{} {
	return Message(false, "Invalid Request")
}

func UnAuthorizedMessage() map[string]interface{} {
	return Message(false, "UnAuthorized")
}

func Respond(w http.ResponseWriter, data map[string] interface{})  {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func GetReadableTime(t time.Time) (string) {

	duration := time.Since(t)
	seconds := math.Round(duration.Seconds())
	if seconds <= 59 {

		if seconds <= 1 {
			return "just now"
		}
		return fmt.Sprintf("%d seconds ago", int(seconds))
	}

	minutes := math.Round(duration.Minutes())
	if minutes <= 59 {
		return fmt.Sprintf("%d minutes ago", int(minutes))
	}

	hours := math.Round(duration.Hours())
	if hours <= 24 {
		return fmt.Sprintf("%d hours ago", int(hours))
	}

	days := hours * 30
	if days <= 730 {
		d := days / 24
		return fmt.Sprintf("%d days ago", int(math.Round(d)))
	}
	months := days / 730
	return fmt.Sprintf("%d months ago", int(math.Round(months)))
}

func RandInt(max, min int) int {
	return rand.Intn(max - min) + min
}
