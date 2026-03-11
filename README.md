# VelocityDBGo 🚀

VelocityDBGo is a lightweight, multi-tenant Backend-as-a-Service (BaaS) designed specifically for students and rapid prototyping. It provides a simple, dynamic NoSQL-like JSON document store backed by the power and reliability of PostgreSQL JSONB.

Built with **Go**, **Gin**, and **GORM**, it allows developers to instantly provision virtual databases (Projects) and tables (Collections) without writing backend code or managing SQL migrations.

## Features ✨

*   **Multi-Tenant Architecture:** Sign up, create Projects (virtual databases), and Collections (virtual tables).
*   **Dynamic Schema (JSONB):** Store arbitrary JSON documents. No need to define columns or migrate schemas.
*   **Advanced Querying:** Filter, sort, and paginate data using intuitive URL query parameters or structured JSON queries.
*   **Deep Nested Lookups:** Query deeply nested JSON fields using dot-notation (e.g., `address.city=London`).
*   **Dual Authentication:** 
    *   **Developer JWT:** For managing projects and collections via the dashboard/CLI.
    *   **Project API Keys:** Permanent, secure keys designed to be embedded in client-side frontend apps (React, Vue, etc.) for data access.
*   **CORS Ready:** Pre-configured to allow cross-origin requests from web frontends.
*   **Web Dashboard:** A beautiful, visual interface for managing projects and collections.
*   **Visual Query Builder:** Explore and filter your data visually without writing any code.
*   **AI Developer Prompt:** One-click copy documentation for AI assistants (ChatGPT/Claude).
*   **Interactive API Docs:** Built-in Swagger UI for testing and exploring the API.

---

## Quick Start 🛠️

### Prerequisites
*   [Go](https://golang.org/doc/install) (v1.20+)
*   [Docker](https://docs.docker.com/get-docker/) & Docker Compose (for the database)

### 1. Start the Database
Spin up the local PostgreSQL database using Docker Compose:
```bash
docker-compose up -d
```

### 2. Configure Environment
Create a `.env` file from the example:
```bash
cp .env.example .env
```
*(Optional: Edit the `.env` file to change default ports or the JWT secret).*

### 3. Run the Server
Install dependencies and start the Go server:
```bash
go mod tidy
go run main.go
```
The server will start on `http://localhost:8080`.

---

## Web Dashboard & Data Explorer 🖥️

VelocityDBGo comes with a built-in, sleek web interface that allows students to manage their database visually.

1.  Run the server and visit: **`http://localhost:8080/`**
2.  **Authentication:** Sign up or log in to access your personal workspace.
3.  **Project Management:** Create projects, view their permanent **API Keys**, and manage collections.
4.  **Data Explorer:** Click on any collection to view, filter, and delete documents in a clean table view.
5.  **Visual Filtering:** Use the built-in query builder to search your data without needing to know the API syntax.
6.  **AI Assistant Integration:** Click the **"Copy AI Prompt"** button in the sidebar to get a pre-written prompt that teaches ChatGPT or Claude exactly how to write code for your project.

---

## Interactive Playground (Swagger) 🎮

The easiest way to understand and test the API is through the interactive Swagger UI.

1.  Run the server.
2.  Open your browser to: **`http://localhost:8080/swagger/index.html`**

You can use the Swagger UI to sign up, log in, create projects, and test all data endpoints without using the terminal.

---

## API Overview 📖

VelocityDBGo has two main API groups: the **Platform API** (for developers) and the **Data API** (for client apps).

### Platform API (Requires Developer JWT)
Developers use these endpoints to manage their infrastructure. Authenticate using the `Authorization: Bearer <token>` header.

*   `POST /api/auth/signup` - Create a developer account.
*   `POST /api/auth/login` - Get a JWT token.
*   `POST /api/projects` - Create a new Project (generates an `APIKey`).
*   `GET /api/projects` - List your Projects.
*   `POST /api/projects/:projectId/collections` - Create a new Collection within a Project.
*   `GET /api/projects/:projectId/collections` - List Collections in a Project.

### Data API (Requires Project API Key OR Developer JWT)
Client applications (like a React frontend) use these endpoints to interact with the database. Authenticate using the `X-API-Key: <key>` header.

*   `POST /api/projects/:projectId/data/:collectionName` - Insert a new JSON document.
*   `GET /api/projects/:projectId/data/:collectionName` - Retrieve and query documents.
*   `POST /api/projects/:projectId/data/:collectionName/query` - Advanced structured JSON querying.
*   `GET /api/projects/:projectId/data/:collectionName/:docId` - Get a single document by ID.
*   `PUT /api/projects/:projectId/data/:collectionName/:docId` - Update/Replace a document.
*   `DELETE /api/projects/:projectId/data/:collectionName/:docId` - Delete a document.

---

## Querying Guide 🔍

VelocityDBGo supports powerful querying capabilities directly against the JSON data.

### 1. URL Query Parameters (`GET`)
Perfect for simple filters and direct browser access.

*   **Exact Match:** `?status=active`
*   **Nested Match:** `?address.city=London`
*   **Operators:** Prefix the value with an operator.
    *   `?rating=gt:4` (Greater than)
    *   `?rating=gte:4` (Greater than or equal)
    *   `?price=lt:100` (Less than)
    *   `?price=lte:100` (Less than or equal)
    *   `?status=neq:closed` (Not equal)
    *   `?name=like:pizza` (Fuzzy search, case-insensitive)
*   **Sorting:** `?sort=rating:desc` or `?sort=name:asc` (Supports nested fields: `?sort=founder.age:desc`)
*   **Pagination:** `?limit=10&offset=0`

### 2. Structured JSON Query (`POST .../query`)
Best for complex queries that are too long for URLs or when querying from code.

```json
POST /api/projects/1/data/restaurants/query
{
  "filter": {
    "type": "italian",
    "location.address.city": "New York",
    "rating": "gte:4.5"
  },
  "sort": "rating:desc",
  "limit": 5,
  "offset": 0
}
```

---

## Built With
*   [Gin](https://gin-gonic.com/) - Web Framework
*   [GORM](https://gorm.io/) - ORM
*   [PostgreSQL](https://www.postgresql.org/) - Core Database
*   [Swaggo](https://github.com/swaggo/swag) - Swagger Documentation Generation
*   **Built with no stress for final year projects 🎓**

---

## License 📜

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
