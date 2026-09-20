package bootstrap

import domaindiscount "github.com/boms/backend/internal/domain/discount"

// Catalog seed data — development only. Prices are USD cents, matching the storefront formatter.
// Slugs are the identity used for idempotent re-runs, so they never change once shipped.

type seedCategory struct {
	Slug      string
	Name      string
	SortOrder int32
	IsActive  bool
}

type seedProduct struct {
	Slug         string
	CategorySlug string
	Name         string
	Description  string
	PriceCents   int64
	IsActive     bool
}

type seedComboItem struct {
	ProductSlug string
	Quantity    int32
}

// seedWindow places a combo relative to the seed run: live, not open yet, or already closed.
type seedWindow string

const (
	windowLive     seedWindow = "live"
	windowUpcoming seedWindow = "upcoming"
	windowClosed   seedWindow = "closed"
)

type seedCombo struct {
	Slug string
	Name string
	// DiscountPercent prices the bundle below the sum of its items (the house rule is 3-8%).
	DiscountPercent int64
	Window          seedWindow
	IsActive        bool
	Items           []seedComboItem
}

var seedCategories = []seedCategory{
	{Slug: "birthday-cakes", Name: "Birthday Cakes", SortOrder: 10, IsActive: true},
	{Slug: "sponge-and-slices", Name: "Sponge & Slices", SortOrder: 20, IsActive: true},
	{Slug: "mini-pastries", Name: "Mini Pastries", SortOrder: 30, IsActive: true},
	{Slug: "croissants", Name: "Croissants & Laminated", SortOrder: 40, IsActive: true},
	{Slug: "cookies", Name: "Cookies", SortOrder: 50, IsActive: true},
	{Slug: "cup-desserts", Name: "Cup Desserts", SortOrder: 60, IsActive: true},
	{Slug: "seasonal", Name: "Seasonal", SortOrder: 70, IsActive: false},
}

var seedProducts = []seedProduct{
	// Birthday cakes — one product per size, since a product carries a single price.
	{Slug: "bento-cake-two-servings", CategorySlug: "birthday-cakes", Name: "Bento Cake, Two Servings", Description: "A 10 cm cake in a takeaway box, piped to order with a short message. Ready in two hours.", PriceCents: 1200, IsActive: true},
	{Slug: "strawberry-shortcake-16", CategorySlug: "birthday-cakes", Name: "Strawberry Shortcake 16 cm", Description: "Vanilla sponge layered with lightly whipped cream and fresh strawberries. Serves four to six.", PriceCents: 2800, IsActive: true},
	{Slug: "strawberry-shortcake-20", CategorySlug: "birthday-cakes", Name: "Strawberry Shortcake 20 cm", Description: "The house strawberry cake in the party size. Serves eight to ten.", PriceCents: 3800, IsActive: true},
	{Slug: "korean-cream-cake-16", CategorySlug: "birthday-cakes", Name: "Korean Cream Cake 16 cm", Description: "Minimalist finish in ivory buttercream with a single piped line of colour. Serves four to six.", PriceCents: 3200, IsActive: true},
	{Slug: "chocolate-fudge-cake-16", CategorySlug: "birthday-cakes", Name: "Chocolate Fudge Cake 16 cm", Description: "Dark chocolate sponge, fudge ganache, cocoa nib crown. Serves four to six.", PriceCents: 3000, IsActive: true},
	{Slug: "tiramisu-whole-cake-16", CategorySlug: "birthday-cakes", Name: "Tiramisu Whole Cake 16 cm", Description: "Espresso-soaked savoiardi under mascarpone cream, dusted at the counter. Serves four to six.", PriceCents: 3400, IsActive: true},

	// Sponge and sliced cakes.
	{Slug: "salted-egg-sponge-cake", CategorySlug: "sponge-and-slices", Name: "Salted Egg Sponge Cake", Description: "Cotton-soft sponge with salted egg custard and pork floss.", PriceCents: 1800, IsActive: true},
	{Slug: "sponge-roll-fresh-cream", CategorySlug: "sponge-and-slices", Name: "Sponge Roll with Fresh Cream", Description: "Rolled vanilla sponge with unsweetened whipped cream. Sold whole.", PriceCents: 1600, IsActive: true},
	{Slug: "taiwanese-cheese-sponge", CategorySlug: "sponge-and-slices", Name: "Taiwanese Cheese Sponge", Description: "Jiggly cheese sponge baked in a water bath, best on the day.", PriceCents: 2000, IsActive: true},
	{Slug: "tiramisu-slice", CategorySlug: "sponge-and-slices", Name: "Tiramisu Slice", Description: "A single serving of the house tiramisu, cut and boxed.", PriceCents: 520, IsActive: true},
	{Slug: "matcha-latte-slice", CategorySlug: "sponge-and-slices", Name: "Matcha Latte Slice", Description: "Ceremonial matcha sponge with milk cream between every layer.", PriceCents: 520, IsActive: true},
	{Slug: "hojicha-cheesecake-slice", CategorySlug: "sponge-and-slices", Name: "Hojicha Cheesecake Slice", Description: "Baked cheesecake infused with roasted hojicha, on a biscuit base.", PriceCents: 540, IsActive: true},

	// Mini pastries.
	{Slug: "cream-puff", CategorySlug: "mini-pastries", Name: "Cream Puff", Description: "The house choux, filled to order with vanilla custard.", PriceCents: 300, IsActive: true},
	{Slug: "matcha-cream-puff", CategorySlug: "mini-pastries", Name: "Matcha Cream Puff", Description: "Choux shell with matcha custard and a dusting of green tea.", PriceCents: 350, IsActive: true},
	{Slug: "madeleine-box-of-six", CategorySlug: "mini-pastries", Name: "Madeleine, Box of Six", Description: "Brown butter madeleines with a lemon zest finish.", PriceCents: 900, IsActive: true},
	{Slug: "financier-box-of-six", CategorySlug: "mini-pastries", Name: "Financier, Box of Six", Description: "Almond financiers baked dark at the edges, soft in the middle.", PriceCents: 1000, IsActive: true},
	{Slug: "canele", CategorySlug: "mini-pastries", Name: "Canele", Description: "Caramelised crust, custard centre, baked in copper moulds.", PriceCents: 350, IsActive: true},
	{Slug: "mini-egg-tart", CategorySlug: "mini-pastries", Name: "Mini Egg Tart", Description: "Flaky shell with a silky egg custard, two bites.", PriceCents: 250, IsActive: true},
	{Slug: "vanilla-cupcake", CategorySlug: "mini-pastries", Name: "Vanilla Cupcake", Description: "Vanilla bean sponge under a swirl of Swiss meringue buttercream.", PriceCents: 350, IsActive: true},
	{Slug: "berry-tartlet", CategorySlug: "mini-pastries", Name: "Berry Tartlet", Description: "Sable shell, vanilla creme, seasonal berries.", PriceCents: 450, IsActive: true},

	// Croissants and laminated doughs.
	{Slug: "butter-croissant", CategorySlug: "croissants", Name: "Butter Croissant", Description: "Three days of lamination with French butter.", PriceCents: 380, IsActive: true},
	{Slug: "almond-croissant", CategorySlug: "croissants", Name: "Almond Croissant", Description: "Yesterday croissant soaked in syrup and filled with frangipane.", PriceCents: 450, IsActive: true},
	{Slug: "croffle-maple", CategorySlug: "croissants", Name: "Croffle with Maple", Description: "Croissant dough pressed in a waffle iron, served with maple.", PriceCents: 400, IsActive: true},
	{Slug: "salted-egg-croissant", CategorySlug: "croissants", Name: "Salted Egg Croissant", Description: "Croissant filled with salted egg custard and finished with floss.", PriceCents: 480, IsActive: true},
	{Slug: "cheese-danish", CategorySlug: "croissants", Name: "Cheese Danish", Description: "Laminated square with cream cheese and a lemon glaze.", PriceCents: 420, IsActive: true},

	// Cookies.
	{Slug: "chocolate-chip-cookie", CategorySlug: "cookies", Name: "Chocolate Chip Cookie", Description: "Thick bakery-style cookie with dark chocolate chunks and flaky salt.", PriceCents: 320, IsActive: true},
	{Slug: "matcha-white-chocolate-cookie", CategorySlug: "cookies", Name: "Matcha White Chocolate Cookie", Description: "Matcha dough with white chocolate, baked soft in the centre.", PriceCents: 340, IsActive: true},
	{Slug: "butter-cookie-tin", CategorySlug: "cookies", Name: "Butter Cookie Tin", Description: "Assorted butter cookies in a keepsake tin. A gifting staple.", PriceCents: 1400, IsActive: true},
	{Slug: "double-chocolate-cookie", CategorySlug: "cookies", Name: "Double Chocolate Cookie", Description: "Cocoa dough with two kinds of chocolate. Off the menu this season.", PriceCents: 320, IsActive: false},

	// Cup desserts.
	{Slug: "tiramisu-cup", CategorySlug: "cup-desserts", Name: "Tiramisu Cup", Description: "The house tiramisu in a sealed cup, made for the ride home.", PriceCents: 480, IsActive: true},
	{Slug: "mango-mousse-cup", CategorySlug: "cup-desserts", Name: "Mango Mousse Cup", Description: "Mango mousse over a sponge base with fresh mango cubes.", PriceCents: 450, IsActive: true},
	{Slug: "creme-caramel-cup", CategorySlug: "cup-desserts", Name: "Creme Caramel Cup", Description: "Steamed custard under a dark caramel, served cold.", PriceCents: 380, IsActive: true},
	{Slug: "almond-jelly-lychee-cup", CategorySlug: "cup-desserts", Name: "Almond Jelly and Lychee Cup", Description: "Almond milk jelly with lychee and toasted almond slices.", PriceCents: 420, IsActive: true},

	// Seasonal — the category is off-menu, so these sit behind it.
	{Slug: "mooncake-gift-box", CategorySlug: "seasonal", Name: "Mooncake Gift Box, Four Pieces", Description: "Four baked mooncakes boxed for the mid-autumn table.", PriceCents: 3600, IsActive: true},
	{Slug: "christmas-yule-log", CategorySlug: "seasonal", Name: "Christmas Yule Log", Description: "Chocolate sponge rolled with chestnut cream, finished as a log.", PriceCents: 2800, IsActive: true},
}

var seedCombos = []seedCombo{
	{
		Slug: "bento-birthday-set", Name: "Bento Birthday Set", DiscountPercent: 6, Window: windowLive, IsActive: true,
		Items: []seedComboItem{
			{ProductSlug: "bento-cake-two-servings", Quantity: 1},
			{ProductSlug: "cream-puff", Quantity: 4},
			{ProductSlug: "tiramisu-cup", Quantity: 1},
		},
	},
	{
		Slug: "morning-pastry-box", Name: "Morning Pastry Box", DiscountPercent: 5, Window: windowLive, IsActive: true,
		Items: []seedComboItem{
			{ProductSlug: "butter-croissant", Quantity: 2},
			{ProductSlug: "almond-croissant", Quantity: 1},
			{ProductSlug: "madeleine-box-of-six", Quantity: 1},
		},
	},
	{
		Slug: "matcha-lovers-set", Name: "Matcha Lovers Set", DiscountPercent: 7, Window: windowUpcoming, IsActive: true,
		Items: []seedComboItem{
			{ProductSlug: "matcha-latte-slice", Quantity: 1},
			{ProductSlug: "matcha-cream-puff", Quantity: 2},
			{ProductSlug: "matcha-white-chocolate-cookie", Quantity: 1},
		},
	},
	{
		Slug: "afternoon-tea-for-two", Name: "Afternoon Tea for Two", DiscountPercent: 4, Window: windowClosed, IsActive: true,
		Items: []seedComboItem{
			{ProductSlug: "berry-tartlet", Quantity: 2},
			{ProductSlug: "canele", Quantity: 2},
			{ProductSlug: "mango-mousse-cup", Quantity: 2},
		},
	},
	{
		Slug: "family-weekend-box", Name: "Family Weekend Box", DiscountPercent: 8, Window: windowLive, IsActive: false,
		Items: []seedComboItem{
			{ProductSlug: "sponge-roll-fresh-cream", Quantity: 1},
			{ProductSlug: "butter-croissant", Quantity: 4},
			{ProductSlug: "chocolate-chip-cookie", Quantity: 2},
		},
	},
}

type seedDiscountCode struct {
	Code             string
	DiscountType     domaindiscount.Type
	Value            int64
	MinOrderCents    *int64
	MaxUses          *int32
	MaxDiscountCents *int64
	UsedCount        int32
	Window           seedWindow
	IsActive         bool
}

func cents(value int64) *int64 { return &value }

func uses(value int32) *int32 { return &value }

var seedDiscountCodes = []seedDiscountCode{
	// Percent off with a cap, the everyday welcome offer.
	{Code: "WELCOME10", DiscountType: domaindiscount.TypePercent, Value: 10, MinOrderCents: cents(2000), MaxDiscountCents: cents(500), Window: windowLive, IsActive: true},
	// Flat amount on a larger basket.
	{Code: "CHOUX5", DiscountType: domaindiscount.TypeFixedCents, Value: 500, MinOrderCents: cents(3000), Window: windowLive, IsActive: true},
	// Capped percent that is nearly out of uses, so the manager list shows a used counter.
	{Code: "BIRTHDAY15", DiscountType: domaindiscount.TypePercent, Value: 15, MinOrderCents: cents(5000), MaxUses: uses(10), MaxDiscountCents: cents(1200), UsedCount: 8, Window: windowLive, IsActive: true},
	// Opens next week.
	{Code: "EARLYBIRD8", DiscountType: domaindiscount.TypePercent, Value: 8, MinOrderCents: cents(1500), MaxDiscountCents: cents(400), Window: windowUpcoming, IsActive: true},
	// Already expired.
	{Code: "SUMMER20", DiscountType: domaindiscount.TypePercent, Value: 20, MinOrderCents: cents(4000), MaxDiscountCents: cents(1000), Window: windowClosed, IsActive: true},
	// In date but switched off by the manager.
	{Code: "STAFFTREAT", DiscountType: domaindiscount.TypeFixedCents, Value: 1000, MinOrderCents: cents(6000), Window: windowLive, IsActive: false},
}
