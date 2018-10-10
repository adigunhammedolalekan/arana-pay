package app

import (
	"golang.org/x/time/rate"
	"time"
	"sync"
	"github.com/gin-gonic/gin"
	"strings"
)

func init() {
	go cleanUp() //Start a background goroutine to cleanup users that hasn't send requests for the past 5minutes
}

var (
	visitors = make(map[string] *Visitor) //Map of Visitors and their ApiKey
	mu sync.Mutex //Lock
)

//A visitor that contains its own Limiter. LastSeen to keep track of the last time this visitor sent a request
type Visitor struct {
	limiter *rate.Limiter
	LastSeen time.Time
}

//Add a visitor to the Map.
func addVisitor(key string) *rate.Limiter {

	l := rate.NewLimiter(5, 5)
	mu.Lock()
	v := &Visitor{
		l, time.Now(),
	}
	visitors[key] = v
	mu.Unlock()

	return l
}

//Get Visitor from Map. Creates it if not exists
func getVisitor(key string) *rate.Limiter {

	mu.Lock()
	v, ok := visitors[key]
	if !ok {
		mu.Unlock()
		return addVisitor(key)
	}

	v.LastSeen = time.Now()
	mu.Unlock()
	return v.limiter
}

//Keep map sane and efficient. Check map every minute and Remove users with LastSeen > 5mins
func cleanUp() {

	for {

		time.Sleep(time.Minute)
		mu.Lock()

		for key, visitor := range visitors {

			if r := time.Now().Sub(visitor.LastSeen); r > 5*time.Minute {
				delete(visitors, key)
			}
		}

		mu.Unlock()
	}
}

func RateLimiterMiddleWare() gin.HandlerFunc {

	return func(c *gin.Context) {

		headerValue := c.Request.Header.Get("Authorization")
		splitted := strings.Split(headerValue, " ") //Value is in form 'Bearer KEY'
		if len(splitted) < 2 {
			c.AbortWithStatusJSON(403, gin.H{"status" : false, "message" : "UnAuthorized"})
			return
		}

		key := splitted[1]
		lim := getVisitor(key)
		if !lim.Allow() {
			c.AbortWithStatusJSON(429, gin.H{"status" : false, "message" : "Too many request"})
			return
		}

		c.Next()
	}
}