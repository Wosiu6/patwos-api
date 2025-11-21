# Patwos API

A Go-based REST API with user authentication and comment management functionality, designed to work with Traefik and Docker to support my personal website/blog.

## API Endpoints

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

### Health Check

- `GET /health` - API health check

### Testing

```bash
go test ./...
```

## Author

Wosiu6
