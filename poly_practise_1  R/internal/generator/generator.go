package generator

import (
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"math/rand"
	"poly_practice_1/config"
	"time"
)

type Generator interface {
	Event() kafka.Message
}

type DefaultGenerator struct {
	userIDs      []string
	productTypes []string
	sources      []string
	brands       []string
	categories   []string
	locations    []string
	statuses     []string
	currencies   []string
	environments []string
	versions     []string
}

func New(cfg *config.ProducerConfig) Generator {
	userIDs := make([]string, cfg.UserCount)
	for i := 0; i < cfg.UserCount; i++ {
		userIDs[i] = fmt.Sprintf("user-%d", i)
	}

	return &DefaultGenerator{
		userIDs:      userIDs,
		productTypes: []string{"electronics", "books", "clothing", "furniture"},
		sources:      []string{"web", "mobile", "api"},
		brands:       []string{"Apple", "Samsung", "Nike", "Adidas", "IKEA", "Sony", "LG", "Dell", "HP", "Asus"},
		categories:   []string{"smartphones", "laptops", "tablets", "accessories", "sports", "home", "office", "gaming"},
		locations:    []string{"warehouse-1", "warehouse-2", "store-1", "store-2", "distribution-center"},
		statuses:     []string{"in_stock", "low_stock", "out_of_stock", "discontinued"},
		currencies:   []string{"USD", "EUR", "GBP", "RUB"},
		environments: []string{"production", "staging", "development"},
		versions:     []string{"v1.0", "v1.1", "v2.0", "v2.1"},
	}
}

func (g *DefaultGenerator) Event() kafka.Message {
	event := g.randomEvent()
	valueBytes, _ := json.Marshal(event)

	return kafka.Message{
		Key:   []byte(fmt.Sprintf("item-%d", event.ItemID)),
		Value: valueBytes,
		Headers: []kafka.Header{
			{Key: "auth_user_id", Value: []byte(g.randomUser())},
			{Key: "product_type", Value: []byte(g.randomProductType())},
			{Key: "trace_id", Value: []byte(g.randomTraceID())},
			{Key: "timestamp", Value: []byte(time.Now().UTC().Format(time.RFC3339))},
			{Key: "source", Value: []byte(g.randomSource())},
			{Key: "category", Value: []byte(event.Category)},
			{Key: "brand", Value: []byte(event.Brand)},
			{Key: "currency", Value: []byte(event.Pricing.Currency)},
			{Key: "environment", Value: []byte(event.Metadata.Environment)},
			{Key: "version", Value: []byte(event.Metadata.Version)},
			{Key: "batch_id", Value: []byte(event.Metadata.BatchID)},
			{Key: "content_type", Value: []byte("application/json")},
			{Key: "encoding", Value: []byte("utf-8")},
		},
	}
}

type MyEvent struct {
	ItemID      int     `json:"item_id"`
	Price       float64 `json:"price"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Brand       string  `json:"brand"`
	SKU         string  `json:"sku"`
	Weight      float64 `json:"weight"`
	Dimensions  struct {
		Length float64 `json:"length"`
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
	} `json:"dimensions"`
	Tags       []string          `json:"tags"`
	Attributes map[string]string `json:"attributes"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
	Inventory  struct {
		Quantity int    `json:"quantity"`
		Location string `json:"location"`
		Status   string `json:"status"`
	} `json:"inventory"`
	Pricing struct {
		BasePrice    float64 `json:"base_price"`
		SalePrice    float64 `json:"sale_price"`
		Currency     string  `json:"currency"`
		DiscountRate float64 `json:"discount_rate"`
	} `json:"pricing"`
	Metadata struct {
		Source      string `json:"source"`
		Version     string `json:"version"`
		Environment string `json:"environment"`
		BatchID     string `json:"batch_id"`
	} `json:"metadata"`
}

func (g *DefaultGenerator) randomUser() string {
	return g.userIDs[rand.Intn(len(g.userIDs))]
}

func (g *DefaultGenerator) randomProductType() string {
	return g.productTypes[rand.Intn(len(g.productTypes))]
}

func (g *DefaultGenerator) randomSource() string {
	return g.sources[rand.Intn(len(g.sources))]
}

func (g *DefaultGenerator) randomTraceID() string {
	return fmt.Sprintf("trace-%d", rand.Int63())
}

func (g *DefaultGenerator) randomEvent() MyEvent {
	basePrice := rand.Float64()*10000 + 100
	discountRate := rand.Float64() * 0.3
	salePrice := basePrice * (1 - discountRate)

	tags := make([]string, rand.Intn(5)+1)
	for i := range tags {
		tags[i] = fmt.Sprintf("tag-%d", rand.Intn(20))
	}

	attributes := map[string]string{
		"color":    fmt.Sprintf("color-%d", rand.Intn(10)),
		"size":     fmt.Sprintf("size-%d", rand.Intn(5)),
		"material": fmt.Sprintf("material-%d", rand.Intn(8)),
		"warranty": fmt.Sprintf("%d-years", rand.Intn(5)+1),
		"shipping": fmt.Sprintf("shipping-%d", rand.Intn(3)),
	}

	now := time.Now().UTC()

	return MyEvent{
		ItemID:      rand.Intn(100000),
		Price:       basePrice,
		Name:        fmt.Sprintf("Product-%d", rand.Intn(1000)),
		Category:    g.categories[rand.Intn(len(g.categories))],
		Description: fmt.Sprintf("High-quality %s product with excellent features", g.categories[rand.Intn(len(g.categories))]),
		Brand:       g.brands[rand.Intn(len(g.brands))],
		SKU:         fmt.Sprintf("SKU-%d-%d", rand.Intn(1000), rand.Intn(1000)),
		Weight:      rand.Float64()*10 + 0.1,
		Dimensions: struct {
			Length float64 `json:"length"`
			Width  float64 `json:"width"`
			Height float64 `json:"height"`
		}{
			Length: rand.Float64()*100 + 1,
			Width:  rand.Float64()*50 + 1,
			Height: rand.Float64()*30 + 1,
		},
		Tags:       tags,
		Attributes: attributes,
		CreatedAt:  now.Format(time.RFC3339),
		UpdatedAt:  now.Format(time.RFC3339),
		Inventory: struct {
			Quantity int    `json:"quantity"`
			Location string `json:"location"`
			Status   string `json:"status"`
		}{
			Quantity: rand.Intn(1000) + 1,
			Location: g.locations[rand.Intn(len(g.locations))],
			Status:   g.statuses[rand.Intn(len(g.statuses))],
		},
		Pricing: struct {
			BasePrice    float64 `json:"base_price"`
			SalePrice    float64 `json:"sale_price"`
			Currency     string  `json:"currency"`
			DiscountRate float64 `json:"discount_rate"`
		}{
			BasePrice:    basePrice,
			SalePrice:    salePrice,
			Currency:     g.currencies[rand.Intn(len(g.currencies))],
			DiscountRate: discountRate,
		},
		Metadata: struct {
			Source      string `json:"source"`
			Version     string `json:"version"`
			Environment string `json:"environment"`
			BatchID     string `json:"batch_id"`
		}{
			Source:      g.sources[rand.Intn(len(g.sources))],
			Version:     g.versions[rand.Intn(len(g.versions))],
			Environment: g.environments[rand.Intn(len(g.environments))],
			BatchID:     fmt.Sprintf("batch-%d", rand.Intn(10000)),
		},
	}
}
