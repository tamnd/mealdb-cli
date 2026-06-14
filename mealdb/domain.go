package mealdb

import (
	"context"
	"time"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes mealdb as a kit Domain driver.
//
// A multi-domain host (ant) enables it with a single blank import:
//
//	import _ "github.com/tamnd/mealdb-cli/mealdb"
//
// The same Domain also builds the standalone mealdb binary (see cli.NewApp).
func init() { kit.Register(Domain{}) }

// Domain is the mealdb driver.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "mealdb",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "mealdb",
			Short:  "Meal and recipe search from TheMealDB",
			Long: `mealdb fetches meal recipes from TheMealDB (themealdb.com) public API.
No API key required. Supports meal search, random picks, category listing,
and cuisine area listing.`,
			Site: Host,
			Repo: "https://github.com/tamnd/mealdb-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// search: find meals by name
	kit.Handle(app, kit.OpMeta{
		Name:    "search",
		Group:   "read",
		List:    true,
		Summary: "Search meals by name",
		Args:    []kit.Arg{{Name: "name", Help: "meal name to search for"}},
	}, searchOp)

	// random: one random meal
	kit.Handle(app, kit.OpMeta{
		Name:    "random",
		Group:   "read",
		Single:  true,
		Summary: "Fetch a random meal",
	}, randomOp)

	// get: fetch meal by ID
	kit.Handle(app, kit.OpMeta{
		Name:    "get",
		Group:   "read",
		Single:  true,
		Summary: "Fetch a meal by ID",
		Args:    []kit.Arg{{Name: "id", Help: "meal ID"}},
	}, getOp)

	// filter: filter meals by category or area
	kit.Handle(app, kit.OpMeta{
		Name:    "filter",
		Group:   "read",
		List:    true,
		Summary: "Filter meals by category or cuisine area",
	}, filterOp)

	// categories: list all meal categories
	kit.Handle(app, kit.OpMeta{
		Name:    "categories",
		Group:   "read",
		List:    true,
		Summary: "List all meal categories",
	}, categoriesOp)

	// areas: list all cuisine areas
	kit.Handle(app, kit.OpMeta{
		Name:    "areas",
		Group:   "read",
		List:    true,
		Summary: "List all cuisine areas/origins",
	}, areasOp)
}

// newClient builds the client from host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type searchInput struct {
	Name   string        `kit:"arg"          help:"meal name to search for"`
	Limit  int           `kit:"flag,inherit" help:"max results"`
	Delay  time.Duration `kit:"flag,inherit" help:"minimum spacing between requests"`
	Client *Client       `kit:"inject"`
}

type randomInput struct {
	Client *Client `kit:"inject"`
}

type getInput struct {
	ID     string  `kit:"arg" help:"meal ID"`
	Client *Client `kit:"inject"`
}

type filterInput struct {
	Category string  `kit:"flag" help:"filter by category (e.g. Seafood, Chicken)"`
	Area     string  `kit:"flag" help:"filter by cuisine area (e.g. Japanese, Italian)"`
	Limit    int     `kit:"flag,inherit" help:"max results"`
	Client   *Client `kit:"inject"`
}

type categoriesInput struct {
	Client *Client `kit:"inject"`
}

type areasInput struct {
	Client *Client `kit:"inject"`
}

// --- handlers ---

func searchOp(ctx context.Context, in searchInput, emit func(Meal) error) error {
	items, err := in.Client.Search(ctx, in.Name, in.Limit)
	if err != nil {
		return mapErr(err)
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

func randomOp(ctx context.Context, in randomInput, emit func(Meal) error) error {
	meal, err := in.Client.Random(ctx)
	if err != nil {
		return mapErr(err)
	}
	return emit(meal)
}

func getOp(ctx context.Context, in getInput, emit func(Meal) error) error {
	meal, err := in.Client.Get(ctx, in.ID)
	if err != nil {
		return mapErr(err)
	}
	return emit(meal)
}

func filterOp(ctx context.Context, in filterInput, emit func(FilterResult) error) error {
	opts := FilterOptions{
		Category: in.Category,
		Area:     in.Area,
		Limit:    in.Limit,
	}
	results, err := in.Client.Filter(ctx, opts)
	if err != nil {
		return mapErr(err)
	}
	for _, r := range results {
		if err := emit(r); err != nil {
			return err
		}
	}
	return nil
}

func categoriesOp(ctx context.Context, in categoriesInput, emit func(Category) error) error {
	cats, err := in.Client.Categories(ctx)
	if err != nil {
		return mapErr(err)
	}
	for _, cat := range cats {
		if err := emit(cat); err != nil {
			return err
		}
	}
	return nil
}

func areasOp(ctx context.Context, in areasInput, emit func(Area) error) error {
	areas, err := in.Client.Areas(ctx)
	if err != nil {
		return mapErr(err)
	}
	for _, area := range areas {
		if err := emit(area); err != nil {
			return err
		}
	}
	return nil
}

// --- Resolver: pure string functions, no network ---

// Classify turns an input into the canonical (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("empty mealdb reference")
	}
	return "meal", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "meal":
		return "https://www.themealdb.com/meal/" + id, nil
	default:
		return "", errs.Usage("mealdb has no resource type %q", uriType)
	}
}

// mapErr converts a library error into the kit error kind.
func mapErr(err error) error {
	return err
}
