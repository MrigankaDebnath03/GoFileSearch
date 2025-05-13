# Product Search API

A lightweight product search microservice built with Go, featuring full-text search capabilities and an in-memory data store.

## Features

- **Full-text Search**: Powered by Bleve search engine for fast and accurate product searches
- **In-memory Storage**: Efficient product data storage using Go maps
- **Concurrent Processing**: Utilizes Go's concurrency features with proper synchronization
- **RESTful API**: Simple and intuitive API endpoints
- **Graceful Shutdown**: Proper server shutdown handling

## Tech Stack

- **Go**: Modern, concurrent programming language
- **Chi Router**: Lightweight and fast HTTP routing
- **Bleve**: Full-text search and indexing library
- **Sync Package**: Go's built-in synchronization primitives for thread safety

## System Architecture

The application follows a simple yet effective architecture:

1. **API Layer**: Chi router handles HTTP requests
2. **Search Engine**: Bleve provides in-memory full-text search capabilities
3. **Data Layer**: Thread-safe in-memory map for product storage

## Getting Started

### Prerequisites

- Go 1.16 or higher
- Git
- Postman (for testing)

### Clone the Repository

```bash
git clone https://github.com/yourusername/product-search-api.git
cd product-search-api
```

### Run the Application

1. Build and run the application:

```bash
go build -o product-api
./product-api
```

2. The server will start on port 8080.

You can also run directly with:

```bash
go run main.go
```

## API Endpoints

### Search Products

```
GET /search?q={query}
```

Search for products by name.

**Parameters:**
- `q` (required): Search query string

**Response:** JSON array of products matching the search query (up to 50 results)

### Add Product

```
POST /products
```

Add a new product.

**Request Body:**
```json
{
  "name": "Product Name",
  "category": "Category Name"
}
```

**Response:** JSON object of the created product including the assigned ID

### Delete Product

```
DELETE /products/{id}
```

Delete a product by ID.

**Parameters:**
- `id` (required): Product ID

**Response:** Empty response with 204 No Content status

## Testing with Postman

### Setting Up Postman

1. Open Postman
2. Create a new Collection named "Product Search API"
3. Add the endpoints described below

### Postman Examples

#### Search for Products

- **Method**: GET
- **URL**: `http://localhost:8080/search?q=mouse`

This will search for all products containing "mouse" in their name.

**Example Response**:
```json
[
  {
    "id": 1,
    "name": "Wireless Mouse 1",
    "category": "Electronics"
  },
  {
    "id": 6,
    "name": "Wireless Mouse 6",
    "category": "Electronics"
  },
  ...
]
```

#### Add a New Product

- **Method**: POST
- **URL**: `http://localhost:8080/products`
- **Headers**: 
  - Content-Type: application/json
- **Body**:
```json
{
  "name": "Mechanical Keyboard",
  "category": "Electronics"
}
```

**Example Response**:
```json
{
  "id": 1001,
  "name": "Mechanical Keyboard",
  "category": "Electronics"
}
```

Note: The application assigns IDs starting from 1001 for newly added products.

#### Delete a Product

- **Method**: DELETE
- **URL**: `http://localhost:8080/products/1001`

This will delete the product with ID 1001.

**Example Response**: Empty response with 204 No Content status

## Implementation Details

### Initial Data

The application is pre-loaded with 1000 products based on 5 template products:
- Wireless Mouse (Electronics)
- Running Shoes (Footwear)
- Gaming Keyboard (Electronics)
- Water Bottle (Home & Kitchen)
- Yoga Mat (Fitness)

### Search Indexing

The application uses Bleve's in-memory index with the following configurations:
- The product name field is indexed for full-text search
- The ID field is indexed as a numeric field
- The category field is stored but not indexed

## Performance Considerations

The application is designed for performance:

1. **In-memory Data Storage**: Fast access to product data without database overhead
2. **In-memory Search Index**: Bleve index for lightning-fast search operations
3. **Thread Safety**: Proper use of mutexes to ensure thread-safe operations
4. **Batched Indexing**: Efficient initial indexing using Bleve's batch operations

## Concurrency Handling

The application uses the following synchronization techniques:

1. **Read-Write Mutex (`sync.RWMutex`)**: Allows concurrent read operations while ensuring exclusive access during writes
2. **Lock Scoping**: Minimizes the critical section to reduce contention
3. **Go Routines**: Separate routine for server startup

## Shutting Down

The application implements graceful shutdown:
- Captures interrupt and termination signals
- Allows a 5-second timeout for ongoing requests to complete
- Properly releases resources

## Advanced Customization

### Modifying Initial Dataset

You can modify the `baseProducts` slice in `main.go` to change the template products or add more templates.

### Changing Search Parameters

The search functionality can be customized by modifying the `searchHandler` function:
- Increase/decrease the limit of search results (currently set to 50)
- Change search fields or add more fields
- Adjust search scoring and sorting

## Troubleshooting

### Server Won't Start

Make sure port 8080 is available. If another application is using this port, you can modify the `Addr` field in the server configuration:

```go
srv := &http.Server{
    Addr:    ":8081", // Change to another port
    Handler: r,
}
```

### Search Not Working

If searches return unexpected results, ensure that:
1. The Bleve index was created successfully (check logs)
2. The search query is properly formatted
3. The product names actually contain the search terms

## License

[MIT License](LICENSE)
