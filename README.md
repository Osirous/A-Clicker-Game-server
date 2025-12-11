# 📦 A Clicker Game – Server  
Backend server for **A Clicker Game**, built in Go with PostgreSQL, JWT authentication, and SQLC.

🖥 **Client Repository:**  
👉 https://github.com/Osirous/A-Clicker-Game  
The client connects directly to this server.

---

# 🧩 Project Overview

This repository contains the **server-side backend** for A Clicker Game.

The backend provides:

- User authentication (register, login)
- Secure JWT session tokens
- Database storage for player data
- REST API endpoints for the client
- SQLC-generated type-safe queries

To actually *play* the game, you also need the **client**, available here:

👉 **https://github.com/Osirous/A-Clicker-Game**

---

# 📁 Project Structure
A-Clicker-Game-server/
├── internal/
│ ├── auth/ # JWT & password hashing
│ └── database/ # SQLC-generated code
├── sql/ # SQL queries
├── main.go # Entry point
├── go.mod
├── go.sum
└── sqlc.yaml


---

# 🔧 Requirements

Make sure you have:

- **Go 1.23+**
- **PostgreSQL 14+**
- Optional: **SQLC** (only needed if regenerating database code)

---

# 🔐 Environment Configuration

The server uses a `.env` file for secrets and database connection.

Create a file named **`.env`** in the project root:

JWT_SECRET=your_jwt_signing_secret_here
DB_URL=postgres://username:password@localhost:5432/clicker?sslmode=disable

### Example Secret
JWT_SECRET=ba922f3c749a48f3b0f7a74e21af91f0790f1c29


> ⚠️ Do NOT commit this file — it is already ignored by `.gitignore`.

---

# 🗄 Database Setup (PostgreSQL)

1. Create the database:

```
createdb clicker
```

2. Apply schema (if included in the repo):
psql clicker < sql/schema.sql

3. Confirm your DB_URL matches your local credentials.

▶️ Running the Server

Install dependencies (only needed the first time):

go mod tidy


Start the backend:

go run .

It will start on localhost:8080 unless changed in the code.

🎮 Running the Client With the Server

1. Clone the client repo:

git clone https://github.com/Osirous/A-Clicker-Game

Run the client in Godot.

Tip: The client will automatically use the server’s endpoints for login, registration, and game data.

🔧 SQLC Usage

If you modify any SQL inside sql/, regenerate the Go DB layer:

sqlc generate

🤝 Contributing

Pull requests are welcome.
