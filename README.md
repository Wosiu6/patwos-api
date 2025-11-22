# Patwos API

A Go-based REST API with user authentication and comment management functionality, designed to work with Traefik and Docker to support my personal website/blog.

## ✨ Features

### Core Functionality
- **User Authentication**
  - Registration with email and username
  - Login with JWT token generation
  - Secure password hashing with bcrypt
  - Token-based authentication middleware

- **Comment System**
  - Create comments on articles
  - Edit own comments
  - Delete own comments
  - View comments by article ID
  - Automatic user association

- **Article Voting** ⭐ NEW
  - Like/dislike articles
  - One vote per user per article
  - Change vote at any time
  - Remove vote
  - Get vote counts (with optional user vote status)
  - Prevents duplicate votes with database constraints

### Architecture
- **SOLID Principles**: Repository pattern, service layer, dependency injection
- **Layered Architecture**: Clear separation of concerns (models, repositories, services, controllers)
- **Interface-Based Design**: Easy to test and extend
- **Production-Ready**: Security, rate limiting, logging, error handling

## 🏗️ Architecture

The API follows SOLID principles with a clean layered architecture:

```
├── models/          → Data models and domain entities
├── repository/      → Data access layer (interfaces + implementations)
├── service/         → Business logic layer
├── controllers/     → HTTP handlers (presentation layer)
├── middleware/      → Cross-cutting concerns (auth, rate limit, CORS)
├── routes/          → Route configuration and dependency injection
├── config/          → Configuration management
└── database/        → Database connection and migrations
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed explanation of SOLID principles implementation.

## 📚 API Endpoints

### Authentication

- `POST /api/v1/auth/register` - Register a new user
  ```json
  {
    "username": "john_doe",
    "email": "john@example.com",
    "password": "securepass123"
  }
  ```

- `POST /api/v1/auth/login` - Login and get JWT token
  ```json
  {
    "email": "john@example.com",
    "password": "securepass123"
  }
  ```

- `GET /api/v1/auth/me` - Get current user (requires authentication)
  - Header: `Authorization: Bearer <token>`

### Comments

- `GET /api/v1/comments/article/:article_id` - Get all comments for an article
- `GET /api/v1/comments/:id` - Get a specific comment
- `POST /api/v1/comments` - Create a new comment (requires authentication)
  ```json
  {
    "content": "Great article!",
    "article_id": "article-slug-123"
  }
  ```

- `PUT/PATCH /api/v1/comments/:id` - Update a comment (requires authentication and ownership)
  ```json
  {
    "content": "Updated comment text"
  }
  ```

- `DELETE /api/v1/comments/:id` - Delete a comment (requires authentication and ownership)

### Article Votes ⭐ NEW

- `POST /api/v1/votes` - Vote on an article (like or dislike, requires authentication)
  ```json
  {
    "article_id": "article-slug-123",
    "vote_type": "like"
  }
  ```

- `GET /api/v1/votes/:article_id` - Get vote counts for an article
  ```json
  {
    "article_id": "article-slug-123",
    "likes": 42,
    "dislikes": 5,
    "user_vote": "like",
    "user_has_voted": true
  }
  ```

- `DELETE /api/v1/votes/:article_id` - Remove your vote (requires authentication)

**Vote Behavior:**
- Each user can only vote once per article
- Users can change their vote (like → dislike or vice versa)
- Users can remove their vote completely
- Vote counts are returned with each vote operation

See [API-DOCUMENTATION.md](API-DOCUMENTATION.md) for complete endpoint documentation with examples.

### Health Check

- `GET /health` - API health check

## Security Features

**⚠️ This API is production-ready with comprehensive security measures:**

### Built-in Security
- ✅ **Rate Limiting**: 100 req/s global, 5 req/min for auth endpoints
- ✅ **JWT Authentication**: Secure token-based auth with 7-day expiration
- ✅ **Password Security**: Bcrypt hashing with proper cost factor
- ✅ **Security Headers**: XSS, clickjacking, MIME-sniffing protection
- ✅ **CORS Protection**: Configurable allowed origins
- ✅ **Input Validation**: Size limits, field validation, SQL injection prevention
- ✅ **HTTPS Enforcement**: Via Traefik with automatic Let's Encrypt
- ✅ **Network Isolation**: Database not exposed to internet

### Critical Production Setup

**Before exposing to internet, you MUST:**

1. **Generate strong JWT secret** (32+ characters):
   ```powershell
   -join ((65..90) + (97..122) + (48..57) | Get-Random -Count 32 | ForEach-Object {[char]$_})
   ```

2. **Set secure database password** (16+ characters)

3. **Configure allowed origins in `.env`**:
   ```env
   ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
   ```

4. **Update CORS in `docker-compose.yml`**:
   ```yaml
   - "traefik.http.middlewares.patwos-cors.headers.accesscontrolalloworiginlist=https://yourdomain.com"
   ```

5. **Set production mode**: `GIN_MODE=release`

6. **Update domain** in docker-compose.yml labels

**📖 See [SECURITY.md](SECURITY.md) for complete security documentation**

### Testing

```bash
go test ./...
```

## Author

Wosiu6
