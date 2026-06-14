package mealdb_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/mealdb-cli/mealdb"
)

const fakeSearchJSON = `{"meals":[{"idMeal":"53000","strMeal":"Chicken Handi","strCategory":"Chicken","strArea":"India","strInstructions":"Cook the chicken with spices until tender.","strMealThumb":"https://www.themealdb.com/images/media/meals/wyxwsp1486979827.jpg","strYoutube":"https://www.youtube.com/watch?v=IO0issT0Rmc","strIngredient1":"Chicken","strMeasure1":"500g","strIngredient2":"Onion","strMeasure2":"2 large","strIngredient3":"Tomato","strMeasure3":"3 medium","strIngredient4":"","strMeasure4":"","strIngredient5":"","strMeasure5":"","strIngredient6":"","strMeasure6":"","strIngredient7":"","strMeasure7":"","strIngredient8":"","strMeasure8":"","strIngredient9":"","strMeasure9":"","strIngredient10":"","strMeasure10":"","strIngredient11":"","strMeasure11":"","strIngredient12":"","strMeasure12":"","strIngredient13":"","strMeasure13":"","strIngredient14":"","strMeasure14":"","strIngredient15":"","strMeasure15":"","strIngredient16":"","strMeasure16":"","strIngredient17":"","strMeasure17":"","strIngredient18":"","strMeasure18":"","strIngredient19":"","strMeasure19":"","strIngredient20":"","strMeasure20":""},{"idMeal":"52772","strMeal":"Teriyaki Chicken Casserole","strCategory":"Chicken","strArea":"Japanese","strInstructions":"Preheat oven to 350.","strMealThumb":"https://www.themealdb.com/images/media/meals/wvpsxx1468256321.jpg","strYoutube":"https://www.youtube.com/watch?v=4aZr5hZXP_s","strIngredient1":"Soy Sauce","strMeasure1":"3/4 cup","strIngredient2":"Water","strMeasure2":"1/2 cup","strIngredient3":"","strMeasure3":"","strIngredient4":"","strMeasure4":"","strIngredient5":"","strMeasure5":"","strIngredient6":"","strMeasure6":"","strIngredient7":"","strMeasure7":"","strIngredient8":"","strMeasure8":"","strIngredient9":"","strMeasure9":"","strIngredient10":"","strMeasure10":"","strIngredient11":"","strMeasure11":"","strIngredient12":"","strMeasure12":"","strIngredient13":"","strMeasure13":"","strIngredient14":"","strMeasure14":"","strIngredient15":"","strMeasure15":"","strIngredient16":"","strMeasure16":"","strIngredient17":"","strMeasure17":"","strIngredient18":"","strMeasure18":"","strIngredient19":"","strMeasure19":"","strIngredient20":"","strMeasure20":""}]}`

const fakeCategoriesJSON = `{"categories":[{"idCategory":"1","strCategory":"Beef","strCategoryThumb":"https://www.themealdb.com/images/category/beef.png","strCategoryDescription":"Beef is the culinary name for meat from cattle."},{"idCategory":"2","strCategory":"Chicken","strCategoryThumb":"https://www.themealdb.com/images/category/chicken.png","strCategoryDescription":"Chicken is a type of domesticated fowl."}]}`

const fakeFilterJSON = `{"meals":[{"idMeal":"52772","strMeal":"Teriyaki Chicken Casserole","strMealThumb":"https://www.themealdb.com/images/media/meals/wvpsxx1468256321.jpg"},{"idMeal":"53049","strMeal":"Chicken Couscous","strMealThumb":"https://www.themealdb.com/images/media/meals/qxytrx1511304021.jpg"}]}`

const fakeAreasJSON = `{"meals":[{"strArea":"American"},{"strArea":"British"},{"strArea":"Canadian"}]}`

func newTestClient(ts *httptest.Server) *mealdb.Client {
	cfg := mealdb.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return mealdb.NewClient(cfg)
}

func TestSearchSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeSearchJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Search(context.Background(), "chicken", 0)
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent header not sent")
	}
}

func TestSearchParsesItems(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeSearchJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Search(context.Background(), "chicken", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}

	got := items[0]
	if got.Name != "Chicken Handi" {
		t.Errorf("items[0].Name = %q, want Chicken Handi", got.Name)
	}
	if got.Category != "Chicken" {
		t.Errorf("items[0].Category = %q, want Chicken", got.Category)
	}
	if got.Area != "India" {
		t.Errorf("items[0].Area = %q, want India", got.Area)
	}
	if len(got.Ingredients) != 3 {
		t.Errorf("len(items[0].Ingredients) = %d, want 3", len(got.Ingredients))
	}
	if items[1].Name != "Teriyaki Chicken Casserole" {
		t.Errorf("items[1].Name = %q, want Teriyaki Chicken Casserole", items[1].Name)
	}
}

func TestSearchLimitRespected(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeSearchJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Search(context.Background(), "chicken", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d, want 1", len(items))
	}
}

func TestSearchRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, fakeSearchJSON)
	}))
	defer ts.Close()

	cfg := mealdb.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 3
	c := mealdb.NewClient(cfg)

	_, err := c.Search(context.Background(), "chicken", 0)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}

func TestGetByID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("i") != "52772" {
			t.Errorf("expected i=52772, got %q", r.URL.Query().Get("i"))
		}
		// Return a single meal
		body := `{"meals":[{"idMeal":"52772","strMeal":"Teriyaki Chicken Casserole","strCategory":"Chicken","strArea":"Japanese","strInstructions":"Preheat oven.","strMealThumb":"https://www.themealdb.com/images/media/meals/wvpsxx1468256321.jpg","strYoutube":"","strIngredient1":"Soy Sauce","strMeasure1":"3/4 cup","strIngredient2":"","strMeasure2":""}]}`
		_, _ = fmt.Fprint(w, body)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	got, err := c.Get(context.Background(), "52772")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "52772" {
		t.Errorf("ID = %q, want 52772", got.ID)
	}
	if got.Name != "Teriyaki Chicken Casserole" {
		t.Errorf("Name = %q, want Teriyaki Chicken Casserole", got.Name)
	}
}

func TestGetNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"meals":null}`)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Get(context.Background(), "99999999")
	if err == nil {
		t.Error("expected error for not-found ID, got nil")
	}
}

func TestFilterByCategory(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("c") == "" {
			t.Errorf("expected c= param, got none")
		}
		_, _ = fmt.Fprint(w, fakeFilterJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	results, err := c.Filter(context.Background(), mealdb.FilterOptions{Category: "Chicken"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Name != "Teriyaki Chicken Casserole" {
		t.Errorf("results[0].Name = %q, want Teriyaki Chicken Casserole", results[0].Name)
	}
	if results[0].ID != "52772" {
		t.Errorf("results[0].ID = %q, want 52772", results[0].ID)
	}
}

func TestFilterNoOptions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeFilterJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Filter(context.Background(), mealdb.FilterOptions{})
	if err == nil {
		t.Error("expected error when no filter options set, got nil")
	}
}

func TestAreasParses(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeAreasJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	areas, err := c.Areas(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(areas) != 3 {
		t.Fatalf("len(areas) = %d, want 3", len(areas))
	}
	if areas[0].Name != "American" {
		t.Errorf("areas[0].Name = %q, want American", areas[0].Name)
	}
}

func TestCategoriesParsesItems(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeCategoriesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Categories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Name != "Beef" {
		t.Errorf("items[0].Name = %q, want Beef", items[0].Name)
	}
	if items[1].Name != "Chicken" {
		t.Errorf("items[1].Name = %q, want Chicken", items[1].Name)
	}
}
