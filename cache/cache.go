package cache

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/go-redis/redis/v8"
)

// Cache defines the structure for the cache
type Cache struct {
	Client *redis.Client
	Ctx    context.Context
}

// NewCache initializes a new cache instance with Redis
func NewCache(addr string) *Cache {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	// Test connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	return &Cache{
		Client: rdb,
		Ctx:    ctx,
	}
}

// Get retrieves a value from Redis
func (c *Cache) Get(key string) (int, bool) {
	val, err := c.Client.Get(c.Ctx, key).Result()
	if err == redis.Nil {
		return 0, false
	} else if err != nil {
		log.Printf("Error getting key from Redis: %v", err)
		return 0, false
	}
	count, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Error converting value to int: %v", err)
		return 0, false
	}
	return count, true
}

// Set adds or updates a value in Redis
func (c *Cache) Set(key string, value int) {
	err := c.Client.Set(c.Ctx, key, value, 0).Err()
	if err != nil {
		log.Printf("Error setting key in Redis: %v", err)
	}
}

// Close closes the Redis client connection
func (c *Cache) Close() error {
	return c.Client.Close()
}

// SetTotalRepos sets the total number of repositories in Redis
func (c *Cache) SetTotalRepos(count int) {
	err := c.Client.Set(c.Ctx, "total_repos", count, 0).Err()
	if err != nil {
		log.Printf("Error setting total_repos in Redis: %v", err)
	}
}

// GetTotalRepos retrieves the total number of repositories from Redis
func (c *Cache) GetTotalRepos() (int, bool) {
	val, err := c.Client.Get(c.Ctx, "total_repos").Result()
	if err == redis.Nil {
		return 0, false
	} else if err != nil {
		log.Printf("Error getting total_repos from Redis: %v", err)
		return 0, false
	}
	count, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Error converting total_repos to int: %v", err)
		return 0, false
	}
	return count, true
}

// SetRepoNames sets the list of repository names in Redis
func (c *Cache) SetRepoNames(repos []string) {
	data, err := json.Marshal(repos)
	if err != nil {
		log.Printf("Error marshaling repo names: %v", err)
		return
	}
	err = c.Client.Set(c.Ctx, "repo_names", data, 0).Err()
	if err != nil {
		log.Printf("Error setting repo_names in Redis: %v", err)
	}
}

// GetRepoNames retrieves the list of repository names from Redis
func (c *Cache) GetRepoNames() ([]string, bool) {
	data, err := c.Client.Get(c.Ctx, "repo_names").Result()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		log.Printf("Error getting repo_names from Redis: %v", err)
		return nil, false
	}
	var repos []string
	err = json.Unmarshal([]byte(data), &repos)
	if err != nil {
		log.Printf("Error unmarshaling repo_names: %v", err)
		return nil, false
	}
	return repos, true
}

// SetKeywords sets the list of keywords in Redis
func (c *Cache) SetKeywords(keywords []string) {
	data, err := json.Marshal(keywords)
	if err != nil {
		log.Printf("Error marshaling keywords: %v", err)
		return
	}
	err = c.Client.Set(c.Ctx, "keywords", data, 0).Err()
	if err != nil {
		log.Printf("Error setting keywords in Redis: %v", err)
	}
}

// GetKeywords retrieves the list of keywords from Redis
func (c *Cache) GetKeywords() ([]string, bool) {
	data, err := c.Client.Get(c.Ctx, "keywords").Result()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		log.Printf("Error getting keywords from Redis: %v", err)
		return nil, false
	}
	var keywords []string
	err = json.Unmarshal([]byte(data), &keywords)
	if err != nil {
		log.Printf("Error unmarshaling keywords: %v", err)
		return nil, false
	}
	return keywords, true
}
