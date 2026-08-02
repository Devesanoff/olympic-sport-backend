#  Olympic Sport Accreditation System (Backend)

High-performance, secure, and scalable RESTful API backend built with **Go (Golang)** for managing athlete accreditation, venue zone access control, meal distribution, and real-time event analytics during large-scale sports events.

---

## 📌 Project Overview

The **Olympic Sport Accreditation System** is designed to solve high-concurrency access control challenges during major athletic tournaments. It ensures offline-ready security using cryptographic signatures, instant zone access validation, real-time tracking via WebSockets, and heavy-duty report generation.

---

## 🛠 Tech Stack & Tools

* **Programming Language:** Go (Golang 1.25)
* **Database:** PostgreSQL (Relational persistence, dynamic filtering)
* **Caching & In-Memory Storage:** Redis (Role-based permissions cache, rate-limiting, counter metrics)
* **Real-time Communication:** Gorilla WebSocket (Concurrent client management)
* **Security & Auth:** JWT (JSON Web Tokens), HMAC-SHA256 (Digital signature for offline QR codes), bcrypt
* **Report & Document Engine:** `github.com/xuri/excelize/v2` (Excel generation), PDF Badge builder
* **Containerization & Infra:** Docker (Multi-stage build), Docker Compose
* **Testing & Benchmarking:** Go standard unit testing, Vegeta (HTTP load testing tool)

---

##  Key Features & Implementation Details

During my internship/development phase, I engineered the backend from scratch adhering to clean architecture principles:

### 1. Authentication & Dynamic RBAC Engine
* Built dynamic **Role-Based Access Control (RBAC)** supporting custom roles (e.g., `SuperAdmin`, `Guard`, `KitchenManager`).
* Integrated **Redis Permission Caching** to evaluate user authorizations in microsecond speed without hammering the primary database.

### 2. Offline-Ready HMAC QR Security
* Implemented **HMAC-SHA256 signed QR codes** for participant badges.
* Allows offline verification at scanning gates, preventing badge tampering and double-entry attacks.

### 3. Real-time Dashboard (WebSockets)
* Engineered a real-time event hub using **Gorilla WebSockets** and Go **goroutines**.
* Broadcasts real-time entry metrics, zone capacity, and food distribution statistics concurrently to monitoring dashboards.

### 4. Zone Access & Meal Control Logic
* Strict **IN/OUT access control** validation per zone.
* Daily **Meal Limit Enforcement** (1 meal per participant per day) with race-condition prevention using atomic Redis operations.

### 5. PDF Badge Builder & Excel Reports Exporter
* Built automatic CR80-standard PDF badge generation with bulk PDF printing support.
* Implemented Excel spreadsheet exports (`.xlsx`) featuring custom headers, auto-fitted columns, status highlighting, and dynamic PostgreSQL query builders (`UNION ALL` for denied access attempt analytics).

---

## ⚡ Performance & Load Testing Results

The system was stress-tested using **Vegeta** against severe traffic bursts to simulate thousands of simultaneous athlete scans at stadium gates.

* **Target Constraint:** < 500 ms latency under high load.
* **Test Load:** 200 concurrent POST requests per second (1,000 total requests) hitting the gate scan endpoint.

### 📊 Benchmark Metrics:
| Metric | Result | Standard |
| :--- | :--- | :--- |
| **P50 Latency (Median)** | **1.51 ms** | < 500 ms (330x faster) |
| **P99 Latency (99th percentile)** | **2.34 ms** | < 500 ms |
| **Max Latency** | **10.67 ms** | - |
| **Success Rate** | **100.00% (200 OK)** | Zero dropped requests |

> **Architecture Secret:** High speed was achieved by decoupling heavy database logging into asynchronous Go goroutines while performing initial access validation directly against Redis.

---

## 💻 Getting Started

### Prerequisites
* [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
* [Go 1.23+](https://golang.org/) (for local development)

### Running with Docker Compose

1. **Clone the repository:**
   ```bash
   git clone [https://github.com/Devesanoff/olympic-sport-backend.git](https://github.com/Devesanoff/olympic-sport-backend.git)
   cd olympic-sport-backend