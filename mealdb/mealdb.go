// Package mealdb is the library behind the mealdb command line:
// the HTTP client, request shaping, and the typed data models for TheMealDB.
//
// The free v1 API uses the static key "1" baked into the base URL path
// (https://www.themealdb.com/api/json/v1/1). No login, no OAuth. The
// Client sets a polite User-Agent, paces requests, and retries transient
// failures (429 and 5xx) with a capped backoff.
package mealdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"sync"
	"time"
)

// Host is the site this client talks to.
const Host = "themealdb.com"

// Config holds all tunable parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns a Config with sensible defaults for the free v1 API.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://www.themealdb.com/api/json/v1/1",
		UserAgent: "Mozilla/5.0 (compatible; mealdb-cli/dev; +https://github.com/tamnd/mealdb-cli)",
		Rate:      300 * time.Millisecond,
		Timeout:   15 * time.Second,
		Retries:   3,
	}
}

// Client talks to TheMealDB over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Search searches meals by name. Returns all matches; trims to limit if > 0.
// Calls GET /search.php?s={encoded_name}.
func (c *Client) Search(ctx context.Context, name string, limit int) ([]Meal, error) {
	u := fmt.Sprintf("%s/search.php?s=%s", c.cfg.BaseURL, neturl.QueryEscape(name))
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp mealsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode search: %w", err)
	}
	items := make([]Meal, 0, len(resp.Meals))
	for i, m := range resp.Meals {
		items = append(items, toMeal(m, i+1))
	}
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}
	return items, nil
}

// Random returns one random meal from the API.
// Calls GET /random.php.
func (c *Client) Random(ctx context.Context) (Meal, error) {
	u := fmt.Sprintf("%s/random.php", c.cfg.BaseURL)
	body, err := c.get(ctx, u)
	if err != nil {
		return Meal{}, err
	}
	var resp mealsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return Meal{}, fmt.Errorf("decode random: %w", err)
	}
	if len(resp.Meals) == 0 {
		return Meal{}, fmt.Errorf("random: no meal returned")
	}
	return toMeal(resp.Meals[0], 1), nil
}

// Lookup returns a meal by ID. Returns an error if the ID is not found.
// Calls GET /lookup.php?i={id}.
func (c *Client) Lookup(ctx context.Context, id string) (Meal, error) {
	u := fmt.Sprintf("%s/lookup.php?i=%s", c.cfg.BaseURL, neturl.QueryEscape(id))
	body, err := c.get(ctx, u)
	if err != nil {
		return Meal{}, err
	}
	var resp mealsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return Meal{}, fmt.Errorf("decode lookup: %w", err)
	}
	if len(resp.Meals) == 0 {
		return Meal{}, fmt.Errorf("meal %s: not found", id)
	}
	return toMeal(resp.Meals[0], 1), nil
}

// FilterOptions controls which filter is applied.
type FilterOptions struct {
	Category string // e.g. "Seafood", "Chicken"
	Area     string // e.g. "Japanese", "Italian"
	Limit    int
}

// Filter returns summary meal records matching the given filter options.
// At least one of Category or Area must be set.
func (c *Client) Filter(ctx context.Context, opts FilterOptions) ([]MealRef, error) {
	var param string
	if opts.Category != "" {
		param = "c=" + neturl.QueryEscape(opts.Category)
	} else if opts.Area != "" {
		param = "a=" + neturl.QueryEscape(opts.Area)
	} else {
		return nil, fmt.Errorf("filter: at least one of category or area must be set")
	}
	u := fmt.Sprintf("%s/filter.php?%s", c.cfg.BaseURL, param)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp filterResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode filter: %w", err)
	}
	results := make([]MealRef, 0, len(resp.Meals))
	for i, m := range resp.Meals {
		results = append(results, MealRef{
			Rank:      i + 1,
			ID:        m.IDMeal,
			Name:      m.StrMeal,
			Thumbnail: m.StrMealThumb,
		})
	}
	if opts.Limit > 0 && opts.Limit < len(results) {
		results = results[:opts.Limit]
	}
	return results, nil
}

// Categories returns all meal categories.
// Calls GET /categories.php.
func (c *Client) Categories(ctx context.Context) ([]Category, error) {
	u := fmt.Sprintf("%s/categories.php", c.cfg.BaseURL)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp categoriesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode categories: %w", err)
	}
	cats := make([]Category, 0, len(resp.Categories))
	for i, cat := range resp.Categories {
		cats = append(cats, Category{
			Rank:        i + 1,
			ID:          cat.IDCategory,
			Name:        cat.StrCategory,
			Description: cat.StrCategoryDescription,
			Thumbnail:   cat.StrCategoryThumb,
		})
	}
	return cats, nil
}

// Areas returns all cuisine areas/origins.
// Calls GET /list.php?a=list.
func (c *Client) Areas(ctx context.Context) ([]Area, error) {
	u := fmt.Sprintf("%s/list.php?a=list", c.cfg.BaseURL)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp areasResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode areas: %w", err)
	}
	areas := make([]Area, 0, len(resp.Meals))
	for i, a := range resp.Meals {
		areas = append(areas, Area{Rank: i + 1, Name: a.StrArea})
	}
	return areas, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return b, err != nil, err
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	return min(time.Duration(attempt)*500*time.Millisecond, 5*time.Second)
}

func toMeal(m rawMeal, rank int) Meal {
	return Meal{
		Rank:         rank,
		ID:           m.IDMeal,
		Name:         m.StrMeal,
		Category:     m.StrCategory,
		Area:         m.StrArea,
		Instructions: m.StrInstructions,
		Thumbnail:    m.StrMealThumb,
		YouTube:      m.StrYoutube,
		Ingredients:  parseIngredients(m),
	}
}
