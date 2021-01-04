package main

import (
	"golang.org/x/sync/singleflight"
	"log"
	"time"
)

func main() {
	var singleSetCache singleflight.Group

	getAndSetCache := func(requestID int, cacheKey string) (string, error) {
		log.Printf("request %v start to get and set cache...", requestID)
		value, _, _ := singleSetCache.Do(cacheKey, func() (ret interface{}, err error) { //do的入参key，可以直接使用缓存的key，这样同一个缓存，只有一个协程会去读DB
			log.Printf("request %v is setting cache...", requestID)
			time.Sleep(3 * time.Second)
			log.Printf("request %v set cache success!", requestID)
			return "VALUE", nil
		})
		return value.(string), nil
	}

	cacheKey := "cacheKey"
	for i := 1; i < 10; i++ { //模拟多个协程同时请求
		go func(requestID int) {
			value, _ := getAndSetCache(requestID, cacheKey)
			log.Printf("request %v get value: %v", requestID, value)
		}(i)
	}
	time.Sleep(20 * time.Second)
}
