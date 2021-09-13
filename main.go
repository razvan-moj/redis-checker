// basic app that polls a managed Redis instance.
package main

import (
  "fmt"
  "log"
  "net/http"
  "os"
  "time"
  "context"
  "github.com/gomodule/redigo/redis"
  "github.com/getsentry/sentry-go"
)

var redisPool *redis.Pool

func incrementHandler(w http.ResponseWriter, r *http.Request) {

  conn := redisPool.Get()
  defer conn.Close()

  counter, err := redis.Int(conn.Do("INCR", "visits"))
  if err != nil {
    http.Error(w, "Error incrementing visitor counter", http.StatusInternalServerError)
    sentry.CaptureException(err)
    return
  }
  fmt.Fprintf(w, "Visitor number: %d", counter)
  log.Printf("Visitor number: %d", counter)

  ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
  defer cancel()
  time.Sleep(time.Second)
  req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://visit-counter:8080", nil)
  if err != nil { log.Printf("%s", err) }
  _, err = http.DefaultClient.Do(req)
  if err != nil { log.Printf("%s", err) }
}

func main() {

  log.SetOutput(os.Stderr)

  err := sentry.Init(sentry.ClientOptions { Dsn: os.Getenv("sentry_dsn"), TracesSampleRate: 1.0, Debug: true, })
  if err != nil { log.Fatalf("sentry.Init: %s", err) }

  redisAddr := os.Getenv("primary_endpoint_address")
  redisToken := os.Getenv("auth_token")
  redisPort := os.Getenv("redis_port")
  if redisPort == "" { redisPort = "6379" }
  redisConn := fmt.Sprintf("%s:%s", redisAddr, redisPort)

  const maxConnections = 1
  redisPool = &redis.Pool{
    MaxIdle: maxConnections,
    Dial: func() (redis.Conn, error) {
      return redis.Dial("tcp", redisConn, redis.DialPassword(redisToken), redis.DialUseTLS(true))
    },
  }

  http.HandleFunc("/", incrementHandler)

  port := os.Getenv("PORT")
  if port == "" { port = "8080" }
  log.Printf("Listening on port %s", port)
  if err := http.ListenAndServe(":"+port, nil); err != nil { log.Fatal(err) }

}
