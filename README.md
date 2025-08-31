# 🚀 Products CRUD API

A **production-ready**, **enterprise-grade** CRUD application for products management built with Go, featuring advanced authentication, caching, session management, and powerful querying capabilities.

## ✨ **Features**

### 🔐 **Authentication & Security**
- **JWT-based Authentication** with access and refresh tokens
- **Session Management** using Redis for multi-device support
- **User Account Isolation** - users can only access their own resources
- **Password Hashing** with bcrypt
- **Input Validation & Sanitization** with SQL injection and XSS protection
- **Secure Token Management** with automatic expiration and refresh
- **Token Blacklisting** - prevents reuse of logged-out tokens
- **Session Validation** - ensures active sessions only
- **Immediate Logout** - tokens become invalid immediately after logout

### 🗄️ **Advanced Data Management**
- **Generic Repository Pattern** for type-safe database operations
- **Advanced Filtering** by price, stock, date ranges, and name
- **Multi-field Sorting** with configurable direction
- **Dual Pagination Systems**:
  - **Offset-based** pagination for traditional navigation
  - **Cursor-based** pagination for high-performance scenarios
- **Product Statistics** with real-time analytics
- **Efficient Query Optimization** with database indexing

### ⚡ **Performance & Scalability**
- **Redis Caching Layer** 
- **Smart Cache Management** 
- **Connection Pooling** 
- **Optimized Database Queries** using GORM and generics

### 🏗️ **Architecture & Design**
- **Clean Architecture** following SOLID principles
- **Layered Design** with clear separation of concerns
- **Dependency Injection** for testable and maintainable code
- **Generic Implementations** reducing code duplication
- **Comprehensive Error Handling** with meaningful error messages

## 🛠️ **Technology Stack**

- **Backend**: Go 1.21+
- **Web Framework**: Gin Gonic
- **Database**: PostgreSQL with GORM ORM
- **Caching**: Redis with JSON serialization
- **Authentication**: JWT with refresh tokens
- **Containerization**: Docker & Docker Compose
- **Validation**: Custom validation with security checks

## 📋 **Prerequisites**

- **Go** 1.21 or higher
- **Docker** and **Docker Compose**
- **PostgreSQL** (via Docker)
- **Redis** (via Docker)


### **Security Features Verified**
- ✅ **User Isolation** - Users can only access their own products
- ✅ **Unauthorized Access Prevention** - Proper 401 responses for invalid requests
- ✅ **Logout Functionality** - Tokens are immediately invalidated
- ✅ **Session Validation** - Only active sessions can access protected endpoints
- ✅ **Token Blacklisting** - Redis-based blacklist prevents token reuse

## 🚀 **Quick Start**

### **Option 1: Docker**

```bash
# Clone the repository
git clone <your-repo-url>
cd products

# Start all services
docker-compose up --build

# The API will be available at http://localhost:8080
```

### **Option 2: Local Development**

```bash
# Clone the repository
git clone <your-repo-url>
cd products

# Copy environment file
cp env.example .env

# Start only the database services
docker-compose up -d postgres redis

# Install dependencies
go mod tidy

# Run the application
go run ./cmd/api
```

## 🔧 **Configuration**

### **Environment Variables**

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=products_db
DB_USER=products_user
DB_PASSWORD=products_password
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# Server Configuration
PORT=8080
```

### **Docker Services**

- **PostgreSQL**: Port 5432
- **Redis**: Port 6379
- **API**: Port 8080

## 📚 **API Endpoints**

### **Authentication**
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/auth/register` | User registration |
| `POST` | `/api/v1/auth/login` | User login |
| `POST` | `/api/v1/auth/refresh` | Refresh access token |
| `POST` | `/api/v1/auth/logout` | Logout from current device |
| `POST` | `/api/v1/auth/logout-all` | Logout from all devices |
| `GET` | `/api/v1/auth/sessions` | Get user's active sessions |

### **Products**
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/products` | Create a new product |
| `GET` | `/api/v1/products` | Get all user's products |
| `GET` | `/api/v1/products/filtered` | Get products with filters, sorting, and pagination |
| `GET` | `/api/v1/products/cursor` | Get products with cursor-based pagination |
| `GET` | `/api/v1/products/stats` | Get product statistics |
| `GET` | `/api/v1/products/:id` | Get a specific product |
| `PUT` | `/api/v1/products/:id` | Update a product |
| `DELETE` | `/api/v1/products/:id` | Delete a product |

## 🔍 **Advanced Querying Examples**

### **Filtering by Price Range**
```bash
GET /api/v1/products/filtered?min_price=20&max_price=100
```

### **Filtering by Stock Level**
```bash
GET /api/v1/products/filtered?min_stock=10
```

### **Date Range Filtering**
```bash
GET /api/v1/products/filtered?created_from=2024-01-01T00:00:00Z&created_to=2024-12-31T23:59:59Z
```

### **Sorting with Pagination**
```bash
GET /api/v1/products/filtered?sort_field=price&sort_direction=desc&page=1&page_size=20
```

### **Cursor-based Pagination**
```bash
GET /api/v1/products/cursor?cursor=uuid&page_size=20&sort_field=created_at&sort_direction=desc
```

## 🧪 **Testing with Postman**

1. **Import Collection**: Import `postman/Products_CRUD_API.postman_collection.json`
2. **Set Environment**: 
   - `base_url`: `http://localhost:8080`
   - `access_token`: Your JWT access token
   - `refresh_token`: Your JWT refresh token
3. **Test Flow**:
   - Register a user
   - Login to get tokens
   - Test CRUD operations
   - Test advanced querying

## 🏗️ **Project Structure**

```
products/
├── cmd/
│   └── api/                    # Application entry point
│       ├── internal/           # Internal packages
│       │   ├── handler/        # HTTP handlers
│       │   ├── router/         # Route definitions
│       │   └── validation/     # Input validation
│       └── main.go            # Main application
├── internal/
│   ├── domain/                # Domain models and interfaces
│   ├── repository/            # Data access layer
│   ├── service/               # Business logic layer
│   └── database/              # Database configuration
├── postman/                   # Postman collection
├── docker-compose.yml         # Docker services
├── Dockerfile                 # Application container
├── Makefile                   # Build automation
└── README.md                  # This file
```

## 🔒 **Security Features**

- **Input Validation**: Comprehensive validation for all inputs
- **SQL Injection Protection**: Pattern-based detection and prevention
- **XSS Protection**: Script tag and JavaScript pattern blocking
- **Input Sanitization**: Removal of dangerous characters
- **JWT Security**: Short-lived access tokens with refresh mechanism
- **Session Management**: Track and control user sessions
- **User Isolation**: Strict resource access control

## 📊 **Performance Features**

- **Redis Caching**
- **Query Optimization**: Efficient database queries with proper indexing
- **Connection Pooling**: Optimized database and Redis connections
- **Smart Pagination**: Handle large datasets efficiently
- **Cache Invalidation**: Automatic cleanup on data changes

## 🚀 **Development Commands**

```bash
# Build the application
make build

# Run tests
make test

# Format code
make fmt

# Lint code
make lint

# Clean build artifacts
make clean

# Run the application
make run

# Docker operations
make docker-build
make docker-up
make docker-down
make docker-logs
```

## 🔄 **Session Management**

The application provides comprehensive session management:

- **Multi-Device Support**: Users can have multiple active sessions
- **Session Tracking**
- **Session Expiration**: Automatic cleanup and renewal
- **Device Control**: Logout from specific devices or all devices

## 💾 **Caching Strategy**

- **Product Data**: 15-30 minutes TTL for frequently accessed data
- **User Data**: 10 minutes TTL for user-specific information
- **Query Results**: 5 minutes TTL for filtered and paginated results
- **Smart Invalidation**: Automatic cache cleanup on data changes
- **Pattern Deletion**: Efficient cache pattern-based cleanup

## 🧪 **Testing Strategy**

- **Postman Collection**: Comprehensive API testing suite collection
