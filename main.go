// basic app that polls a managed Redis instance.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
  "time"
  "io"
	"github.com/gomodule/redigo/redis"
)

var redisPool *redis.Pool

func incrementHandler(w http.ResponseWriter, r *http.Request) {

	conn := redisPool.Get()
	defer conn.Close()

	counter, err := redis.Int(conn.Do("INCR", "visits"))
	if err != nil {
		http.Error(w, "Error incrementing visitor counter", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Visitor number: %d", counter)
	log.Printf("Visitor number: %d", counter)

	time.Sleep(time.Second)
	resp, err := http.Get("http://visit-counter:8080")
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	log.Printf(string(body))

}

func main() {

	log.SetOutput(os.Stderr)

	redisAddr := os.Getenv("primary_endpoint_address")
  redisToken := os.Getenv("auth_token")
  redisPort := os.Getenv("redis_port")
  if redisPort == "" {
    redisPort = "6379"
	}
	redisConn := fmt.Sprintf("%s:%s", redisAddr, redisPort)

	const maxConnections = 10
	redisPool = &redis.Pool{
		MaxIdle: maxConnections,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisConn, redis.DialPassword(redisToken), redis.DialUseTLS(true))
		},
	}

	http.HandleFunc("/", incrementHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}

}
