// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"strconv"
// 	"syscall"
// 	"time"

// 	"github.com/blevesearch/bleve/v2"
// 	"github.com/go-chi/chi/v5"
// )

// type Product struct {
// 	ID       int    `json:"id"`
// 	Name     string `json:"name"`
// 	Category string `json:"category"`
// }

// var baseProducts = []Product{
// 	{ID: 1, Name: "Wireless Mouse", Category: "Electronics"},
// 	{ID: 2, Name: "Running Shoes", Category: "Footwear"},
// 	{ID: 3, Name: "Gaming Keyboard", Category: "Electronics"},
// 	{ID: 4, Name: "Water Bottle", Category: "Home & Kitchen"},
// 	{ID: 5, Name: "Yoga Mat", Category: "Fitness"},
// }

// func main() {
// 	products := generateProducts()

// 	index, err := createIndex(products)
// 	if err != nil {
// 		log.Fatalf("Error creating index: %v", err)
// 	}
// 	defer index.Close()

// 	r := chi.NewRouter()
// 	r.Get("/search", searchHandler(products, index))

// 	srv := &http.Server{
// 		Addr:    ":8080",
// 		Handler: r,
// 	}

// 	go func() {
// 		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			log.Fatalf("Server error: %v", err)
// 		}
// 	}()

// 	log.Println("Server started on :8080")

// 	sigChan := make(chan os.Signal, 1)
// 	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
// 	<-sigChan

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// 	if err := srv.Shutdown(ctx); err != nil {
// 		log.Fatalf("Server shutdown error: %v", err)
// 	}
// 	log.Println("Server gracefully stopped")
// }

// func generateProducts() []Product {
// 	products := make([]Product, 1e6)
// 	baseLen := len(baseProducts)
// 	for i := 0; i < 1000; i++ {
// 		base := baseProducts[i%baseLen]
// 		products[i] = Product{
// 			ID:       i + 1,
// 			Name:     fmt.Sprintf("%s %d", base.Name, i+1),
// 			Category: base.Category,
// 		}
// 	}
// 	return products
// }

// func createIndex(products []Product) (bleve.Index, error) {

// 	mapping := bleve.NewIndexMapping()

// 	productMapping := bleve.NewDocumentMapping()

// 	nameFieldMapping := bleve.NewTextFieldMapping()
// 	productMapping.AddFieldMappingsAt("Name", nameFieldMapping)

// 	idFieldMapping := bleve.NewNumericFieldMapping()
// 	productMapping.AddFieldMappingsAt("ID", idFieldMapping)

// 	categoryFieldMapping := bleve.NewTextFieldMapping()
// 	categoryFieldMapping.Index = false
// 	productMapping.AddFieldMappingsAt("Category", categoryFieldMapping)

// 	mapping.AddDocumentMapping("product", productMapping)
// 	mapping.DefaultAnalyzer = "standard"

// 	index, err := bleve.NewMemOnly(mapping)
// 	if err != nil {
// 		return nil, err
// 	}

// 	batch := index.NewBatch()
// 	for i, p := range products {
// 		err := batch.Index(strconv.Itoa(p.ID), map[string]interface{}{
// 			"ID":       p.ID,
// 			"Name":     p.Name,
// 			"Category": p.Category,
// 		})
// 		if err != nil {
// 			return nil, err
// 		}

// 		if (i+1)%10000 == 0 {
// 			if err := index.Batch(batch); err != nil {
// 				return nil, err
// 			}
// 			batch = index.NewBatch()
// 		}
// 	}

// 	if err := index.Batch(batch); err != nil {
// 		return nil, err
// 	}

// 	return index, nil
// }

// func searchHandler(products []Product, index bleve.Index) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		queryParam := r.URL.Query().Get("q")

// 		matchQuery := bleve.NewMatchQuery(queryParam)
// 		matchQuery.SetField("Name")
// 		searchRequest := bleve.NewSearchRequestOptions(matchQuery, 50, 0, false)

// 		searchResult, err := index.Search(searchRequest)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		var results []Product
// 		for _, hit := range searchResult.Hits {
// 			id, err := strconv.Atoi(hit.ID)
// 			if err != nil {
// 				continue
// 			}
// 			if id < 1 || id > len(products) {
// 				continue
// 			}
// 			product := products[id-1]
// 			results = append(results, product)
// 			if len(results) >= 50 {
// 				break
// 			}
// 		}

// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(results)
// 	}
// }

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/go-chi/chi/v5"
)

type Product struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

var (
	products map[int]Product
	mu       sync.RWMutex
	nextID   int = 1001 // Starting ID for new products
)

var baseProducts = []Product{
	{ID: 1, Name: "Wireless Mouse", Category: "Electronics"},
	{ID: 2, Name: "Running Shoes", Category: "Footwear"},
	{ID: 3, Name: "Gaming Keyboard", Category: "Electronics"},
	{ID: 4, Name: "Water Bottle", Category: "Home & Kitchen"},
	{ID: 5, Name: "Yoga Mat", Category: "Fitness"},
}

func main() {
	// Generate initial dataset
	products = generateProducts()

	// Create and populate Bleve index
	index, err := createIndex()
	if err != nil {
		log.Fatalf("Error creating index: %v", err)
	}
	defer index.Close()

	// Create router and register handlers
	r := chi.NewRouter()
	r.Get("/search", searchHandler(index))
	r.Post("/products", addProductHandler(index))
	r.Delete("/products/{id}", deleteProductHandler(index))

	// Configure server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()
	log.Println("Server started on :8080")

	// Graceful shutdown setup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
	log.Println("Server gracefully stopped")
}

func generateProducts() map[int]Product {
	products := make(map[int]Product, 1e6)
	baseLen := len(baseProducts)

	mu.Lock()
	defer mu.Unlock()

	for i := 0; i < 1000; i++ {
		base := baseProducts[i%baseLen]
		id := i + 1
		products[id] = Product{
			ID:       id,
			Name:     fmt.Sprintf("%s %d", base.Name, id),
			Category: base.Category,
		}
	}
	return products
}

func createIndex() (bleve.Index, error) {
	mapping := bleve.NewIndexMapping()

	productMapping := bleve.NewDocumentMapping()

	nameFieldMapping := bleve.NewTextFieldMapping()
	productMapping.AddFieldMappingsAt("Name", nameFieldMapping)

	idFieldMapping := bleve.NewNumericFieldMapping()
	productMapping.AddFieldMappingsAt("ID", idFieldMapping)

	categoryFieldMapping := bleve.NewTextFieldMapping()
	categoryFieldMapping.Index = false
	productMapping.AddFieldMappingsAt("Category", categoryFieldMapping)

	mapping.AddDocumentMapping("product", productMapping)
	mapping.DefaultAnalyzer = "standard"

	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}

	batch := index.NewBatch()
	mu.RLock()
	defer mu.RUnlock()

	for _, p := range products {
		err := batch.Index(strconv.Itoa(p.ID), map[string]interface{}{
			"ID":       p.ID,
			"Name":     p.Name,
			"Category": p.Category,
		})
		if err != nil {
			return nil, err
		}
	}

	if err := index.Batch(batch); err != nil {
		return nil, err
	}
	return index, nil
}

func searchHandler(index bleve.Index) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParam := r.URL.Query().Get("q")

		matchQuery := bleve.NewMatchQuery(queryParam)
		matchQuery.SetField("Name")
		searchRequest := bleve.NewSearchRequestOptions(matchQuery, 50, 0, false)

		searchResult, err := index.Search(searchRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		mu.RLock()
		defer mu.RUnlock()

		var results []Product
		for _, hit := range searchResult.Hits {
			id, err := strconv.Atoi(hit.ID)
			if err != nil {
				continue
			}
			if product, exists := products[id]; exists {
				results = append(results, product)
				if len(results) >= 50 {
					break
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

func addProductHandler(index bleve.Index) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p Product
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		p.ID = nextID
		nextID++
		products[p.ID] = p

		if err := index.Index(strconv.Itoa(p.ID), p); err != nil {
			delete(products, p.ID)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)
	}
}

func deleteProductHandler(index bleve.Index) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if _, exists := products[id]; !exists {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}

		delete(products, id)
		if err := index.Delete(strconv.Itoa(id)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
