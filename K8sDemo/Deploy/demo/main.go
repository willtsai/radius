package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis"
)

func hello(w http.ResponseWriter, req *http.Request) {
	redisHost := os.Getenv("CONNECTION_REDIS_HOST")
	redisPort := os.Getenv("CONNECTION_REDIS_PORT")
	password := os.Getenv("CONNECTION_REDIS_PASSWORD")
	if password == "" {
		log.Fatal("Failed to fetch password.")
	}
	log.Println("Successfully fetched Redis password.")
	log.Printf("Connecting to Redis at %s:%s...\n", redisHost, redisPort)
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: password, // no password set
		DB:       0,        // use default DB
	})
	value := fmt.Sprintf("%d", time.Now().Unix())
	key := "key-" + value
	if err := rdb.Set(key, value, 10*time.Second).Err(); err != nil {
		panic(err)
	}

	_, _ = w.Write([]byte("Writing key: " + key + " with value: " + value + "\n"))

	val, err := rdb.Get(key).Result()
	if err != nil {
		panic(err)
	} else if val != value {
		panic(fmt.Sprintf("expected %s, saw %s", value, val))
	}
	log.Println("Succesfully wrote a test value to Redis and read it back.")

	_, _ = w.Write([]byte("Retrieved value: " + val))
}

func main() {
	log.Println("Starting demo app...")

	http.HandleFunc("/", hello)
	_ = http.ListenAndServe(":8000", nil)
}
