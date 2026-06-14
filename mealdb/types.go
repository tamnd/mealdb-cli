package mealdb

import "strings"

// Ingredient is one meal ingredient with its measure.
type Ingredient struct {
	Name    string `json:"name"`
	Measure string `json:"measure"`
}

// Meal is one recipe from TheMealDB.
type Meal struct {
	Rank         int          `json:"rank"`
	ID           string       `kit:"id" json:"id"`
	Name         string       `json:"name"`
	Category     string       `json:"category"`
	Area         string       `json:"area"` // cuisine origin
	Instructions string       `json:"instructions"`
	Thumbnail    string       `json:"thumbnail"`
	YouTube      string       `json:"youtube"`
	Ingredients  []Ingredient `json:"ingredients"`
}

// Category is one meal category from TheMealDB.
type Category struct {
	Rank        int    `json:"rank"`
	ID          string `kit:"id" json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
}

// Area is one cuisine area/origin from TheMealDB.
type Area struct {
	Rank int    `json:"rank"`
	Name string `json:"name"`
}

// MealRef is a summary meal reference returned by the filter endpoint.
// It does not include ingredients; use Lookup to fetch full details.
type MealRef struct {
	Rank      int    `json:"rank"`
	ID        string `kit:"id" json:"id"`
	Name      string `json:"name"`
	Thumbnail string `json:"thumbnail"`
}

// --- wire types (unexported, only used for JSON decoding) ---

type rawMeal struct {
	IDMeal          string `json:"idMeal"`
	StrMeal         string `json:"strMeal"`
	StrCategory     string `json:"strCategory"`
	StrArea         string `json:"strArea"`
	StrInstructions string `json:"strInstructions"`
	StrMealThumb    string `json:"strMealThumb"`
	StrYoutube      string `json:"strYoutube"`
	// Ingredients 1-20
	StrIngredient1  string `json:"strIngredient1"`
	StrIngredient2  string `json:"strIngredient2"`
	StrIngredient3  string `json:"strIngredient3"`
	StrIngredient4  string `json:"strIngredient4"`
	StrIngredient5  string `json:"strIngredient5"`
	StrIngredient6  string `json:"strIngredient6"`
	StrIngredient7  string `json:"strIngredient7"`
	StrIngredient8  string `json:"strIngredient8"`
	StrIngredient9  string `json:"strIngredient9"`
	StrIngredient10 string `json:"strIngredient10"`
	StrIngredient11 string `json:"strIngredient11"`
	StrIngredient12 string `json:"strIngredient12"`
	StrIngredient13 string `json:"strIngredient13"`
	StrIngredient14 string `json:"strIngredient14"`
	StrIngredient15 string `json:"strIngredient15"`
	StrIngredient16 string `json:"strIngredient16"`
	StrIngredient17 string `json:"strIngredient17"`
	StrIngredient18 string `json:"strIngredient18"`
	StrIngredient19 string `json:"strIngredient19"`
	StrIngredient20 string `json:"strIngredient20"`
	// Measures 1-20
	StrMeasure1  string `json:"strMeasure1"`
	StrMeasure2  string `json:"strMeasure2"`
	StrMeasure3  string `json:"strMeasure3"`
	StrMeasure4  string `json:"strMeasure4"`
	StrMeasure5  string `json:"strMeasure5"`
	StrMeasure6  string `json:"strMeasure6"`
	StrMeasure7  string `json:"strMeasure7"`
	StrMeasure8  string `json:"strMeasure8"`
	StrMeasure9  string `json:"strMeasure9"`
	StrMeasure10 string `json:"strMeasure10"`
	StrMeasure11 string `json:"strMeasure11"`
	StrMeasure12 string `json:"strMeasure12"`
	StrMeasure13 string `json:"strMeasure13"`
	StrMeasure14 string `json:"strMeasure14"`
	StrMeasure15 string `json:"strMeasure15"`
	StrMeasure16 string `json:"strMeasure16"`
	StrMeasure17 string `json:"strMeasure17"`
	StrMeasure18 string `json:"strMeasure18"`
	StrMeasure19 string `json:"strMeasure19"`
	StrMeasure20 string `json:"strMeasure20"`
}

type mealsResponse struct {
	Meals []rawMeal `json:"meals"`
}

type rawCategory struct {
	IDCategory             string `json:"idCategory"`
	StrCategory            string `json:"strCategory"`
	StrCategoryThumb       string `json:"strCategoryThumb"`
	StrCategoryDescription string `json:"strCategoryDescription"`
}

type categoriesResponse struct {
	Categories []rawCategory `json:"categories"`
}

type areasResponse struct {
	Meals []struct {
		StrArea string `json:"strArea"`
	} `json:"meals"`
}

type filterMeal struct {
	IDMeal       string `json:"idMeal"`
	StrMeal      string `json:"strMeal"`
	StrMealThumb string `json:"strMealThumb"`
}

type filterResponse struct {
	Meals []filterMeal `json:"meals"`
}

// parseIngredients converts the flat strIngredientN / strMeasureN fields of a
// rawMeal into a clean []Ingredient. Slots are filled consecutively: the loop
// stops on the first empty name.
func parseIngredients(m rawMeal) []Ingredient {
	names := [20]string{
		m.StrIngredient1, m.StrIngredient2, m.StrIngredient3,
		m.StrIngredient4, m.StrIngredient5, m.StrIngredient6,
		m.StrIngredient7, m.StrIngredient8, m.StrIngredient9,
		m.StrIngredient10, m.StrIngredient11, m.StrIngredient12,
		m.StrIngredient13, m.StrIngredient14, m.StrIngredient15,
		m.StrIngredient16, m.StrIngredient17, m.StrIngredient18,
		m.StrIngredient19, m.StrIngredient20,
	}
	measures := [20]string{
		m.StrMeasure1, m.StrMeasure2, m.StrMeasure3,
		m.StrMeasure4, m.StrMeasure5, m.StrMeasure6,
		m.StrMeasure7, m.StrMeasure8, m.StrMeasure9,
		m.StrMeasure10, m.StrMeasure11, m.StrMeasure12,
		m.StrMeasure13, m.StrMeasure14, m.StrMeasure15,
		m.StrMeasure16, m.StrMeasure17, m.StrMeasure18,
		m.StrMeasure19, m.StrMeasure20,
	}
	var ings []Ingredient
	for i, name := range names {
		if name == "" {
			break
		}
		ings = append(ings, Ingredient{
			Name:    strings.TrimSpace(name),
			Measure: strings.TrimSpace(measures[i]),
		})
	}
	return ings
}
