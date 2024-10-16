package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mouismail/search/cache"

	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

// Configuration represents the structure of the YAML config file
type Configuration struct {
	OrgName  string   `yaml:"org_name"`
	Keywords []string `yaml:"keywords"`
}

// SearchTask represents a search operation with repo and keyword
type SearchTask struct {
	Repo    string
	Keyword string
}

// Result represents the search result
type Result struct {
	Repo    string
	Keyword string
	Count   int
}

const (
	rateLimit  = 10            // requests per minute
	configFile = "config.yaml" // Path to the YAML config file
)

func main() {
	// Load configuration from YAML file
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Retrieve GitHub Token from environment variable
	githubToken, exists := os.LookupEnv("GITHUB_TOKEN")
	if !exists || githubToken == "" {
		log.Fatalf("GITHUB_TOKEN environment variable is not set")
	}

	// Retrieve Redis address from environment variable
	redisAddr, exists := os.LookupEnv("REDIS_ADDR")
	if !exists || redisAddr == "" {
		log.Fatalf("REDIS_ADDR environment variable is not set")
	}

	// Initialize Redis Cache
	c := cache.NewCache(redisAddr)
	defer func() {
		if err := c.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}()

	ctx := context.Background()

	var allRepos []string
	var keywords []string

	// Initialize GitHub client with OAuth2 authentication
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Attempt to get cached repo names and keywords
	cachedRepos, reposExists := c.GetRepoNames()
	cachedKeywords, keywordsExist := c.GetKeywords()

	if reposExists && keywordsExist {
		allRepos = cachedRepos
		keywords = cachedKeywords
		// Get total repos from cache
		total, totalExists := c.GetTotalRepos()
		if !totalExists {
			total = len(allRepos)
			c.SetTotalRepos(total)
		}
		fmt.Printf("Total Repositories (from cache): %d\n", total)
	} else {
		// Fetch repositories from GitHub
		allRepos, err = fetchAllRepositories(ctx, client, config.OrgName)
		if err != nil {
			log.Fatalf("Error fetching repositories: %v", err)
		}
		fmt.Printf("Total Repositories: %d\n", len(allRepos))
		// Cache the fetched data
		c.SetTotalRepos(len(allRepos))
		c.SetRepoNames(allRepos)
		c.SetKeywords(config.Keywords)
		keywords = config.Keywords
	}

	// Use cached or loaded keywords
	if !keywordsExist {
		keywords = config.Keywords
	}

	// Create channels
	tasks := make(chan SearchTask, len(allRepos)*len(keywords))
	results := make(chan Result, len(allRepos)*len(keywords))

	var wg sync.WaitGroup

	// Rate limiter: rateLimit requests per minute
	limiter := time.Tick(time.Minute / time.Duration(rateLimit))

	// Worker function
	worker := func() {
		defer wg.Done()
		for task := range tasks {
			cacheKey := fmt.Sprintf("%s:%s", task.Repo, task.Keyword)

			// Check cache
			if count, exists := c.Get(cacheKey); exists {
				results <- Result{
					Repo:    task.Repo,
					Keyword: task.Keyword,
					Count:   count,
				}
				continue
			}

			// Respect rate limiting
			<-limiter

			// Perform GitHub code search
			searchQuery := fmt.Sprintf("repo:%s %s", task.Repo, task.Keyword)
			opts := &github.SearchOptions{
				Sort:  "indexed",
				Order: "desc",
				ListOptions: github.ListOptions{
					PerPage: 100,
				},
			}

			searchResult, _, err := client.Search.Code(ctx, searchQuery, opts)
			if err != nil {
				log.Printf("Error searching '%s' in '%s': %v", task.Keyword, task.Repo, err)
				continue
			}

			count := searchResult.GetTotal()

			// Update cache
			c.Set(cacheKey, count)

			// Send result
			results <- Result{
				Repo:    task.Repo,
				Keyword: task.Keyword,
				Count:   count,
			}
		}
	}

	// Start workers
	numWorkers := rateLimit
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go worker()
	}

	// Enqueue tasks
	go func() {
		for _, repo := range allRepos {
			for _, keyword := range keywords {
				tasks <- SearchTask{Repo: repo, Keyword: keyword}
			}
		}
		close(tasks)
	}()

	// Handle graceful shutdown
	go handleGracefulShutdown(&wg, &results)

	// Collect and display results
	for res := range results {
		fmt.Printf("Repo: %s, Keyword: %s, Count: %d\n", res.Repo, res.Keyword, res.Count)
	}
}

// loadConfig reads and parses the YAML configuration file
func loadConfig(filename string) (*Configuration, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Configuration
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// fetchAllRepositories retrieves all repository full names in the specified organization
func fetchAllRepositories(ctx context.Context, client *github.Client, org string) ([]string, error) {
	var allRepos []string
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			allRepos = append(allRepos, repo.GetFullName())
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

// handleGracefulShutdown listens for interrupt signals to gracefully shut down the application
func handleGracefulShutdown(wg *sync.WaitGroup, results *chan Result) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
	fmt.Println("\nReceived interrupt signal. Waiting for ongoing tasks to finish...")

	// Wait for all workers to finish
	wg.Wait()

	// Close the results channel to stop the results collector
	close(*results)

	fmt.Println("Graceful shutdown complete.")
}
