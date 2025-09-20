# Blog Platform API Examples

## Authentication

### Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

Response:
```json
{
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Login User
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

Response:
```json
{
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Posts

### Create Post (Authenticated)
```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "title": "My First Blog Post",
    "content": "This is the content of my first blog post. It contains interesting information about web development."
  }'
```

Response:
```json
{
  "id": 1,
  "title": "My First Blog Post",
  "content": "This is the content of my first blog post. It contains interesting information about web development.",
  "author_id": 1,
  "created_at": "2024-01-15T11:00:00Z",
  "updated_at": "2024-01-15T11:00:00Z"
}
```

### Get All Posts
```bash
curl -X GET "http://localhost:8080/api/v1/posts?limit=10&offset=0"
```

Response:
```json
{
  "posts": [
    {
      "id": 1,
      "title": "My First Blog Post",
      "content": "This is the content of my first blog post...",
      "author_id": 1,
      "created_at": "2024-01-15T11:00:00Z",
      "updated_at": "2024-01-15T11:00:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

### Get Post by ID
```bash
curl -X GET http://localhost:8080/api/v1/posts/1
```

Response:
```json
{
  "id": 1,
  "title": "My First Blog Post",
  "content": "This is the content of my first blog post. It contains interesting information about web development.",
  "author_id": 1,
  "created_at": "2024-01-15T11:00:00Z",
  "updated_at": "2024-01-15T11:00:00Z"
}
```

### Update Post (Authenticated, Author Only)
```bash
curl -X PUT http://localhost:8080/api/v1/posts/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "title": "My Updated Blog Post",
    "content": "This is the updated content of my blog post with new information."
  }'
```

Response:
```json
{
  "id": 1,
  "title": "My Updated Blog Post",
  "content": "This is the updated content of my blog post with new information.",
  "author_id": 1,
  "created_at": "2024-01-15T11:00:00Z",
  "updated_at": "2024-01-15T11:30:00Z"
}
```

### Delete Post (Authenticated, Author Only)
```bash
curl -X DELETE http://localhost:8080/api/v1/posts/1 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

Response: `204 No Content`

## Comments

### Create Comment
```bash
curl -X POST http://localhost:8080/api/v1/posts/1/comments \
  -H "Content-Type: application/json" \
  -d '{
    "author_name": "Jane Smith",
    "content": "Great post! Very informative and well written."
  }'
```

Response:
```json
{
  "id": 1,
  "post_id": 1,
  "author_name": "Jane Smith",
  "content": "Great post! Very informative and well written.",
  "created_at": "2024-01-15T12:00:00Z"
}
```

### Get Comments for Post
```bash
curl -X GET "http://localhost:8080/api/v1/posts/1/comments?limit=10&offset=0"
```

Response:
```json
{
  "comments": [
    {
      "id": 1,
      "post_id": 1,
      "author_name": "Jane Smith",
      "content": "Great post! Very informative and well written.",
      "created_at": "2024-01-15T12:00:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

## Error Responses

### Validation Error (400)
```json
{
  "error": "validation_error",
  "message": "Key: 'RegisterRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"
}
```

### Unauthorized (401)
```json
{
  "error": "unauthorized",
  "message": "Missing or invalid authorization token"
}
```

### Forbidden (403)
```json
{
  "error": "forbidden",
  "message": "You can only modify your own posts"
}
```

### Not Found (404)
```json
{
  "error": "not_found",
  "message": "Post not found"
}
```

### Conflict (409)
```json
{
  "error": "user_exists",
  "message": "User with this email already exists"
}
```

### Internal Server Error (500)
```json
{
  "error": "internal_error",
  "message": "An internal server error occurred"
}
```
