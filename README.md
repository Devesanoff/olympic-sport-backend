# 🏅 Olympic Sport Accreditation System (Backend) - O'zbekcha

Yirik sport musobaqalarida sportchilarni akkreditatsiyadan o'tkazish, hududlarga (zone) kirishni nazorat qilish, ovqatlanish tizimini boshqarish va real vaqtdagi analitikani yuritish uchun **Go (Golang)** tilida yozilgan yuqori unumdorlikka ega, xavfsiz va kengayuvchan RESTful API backend tizimi.

---

## 📌 Loyiha haqida umumiy ma'lumot

**Olympic Sport Accreditation System** — bu yirik turnirlarda bir vaqtning o'zida yuzaga keladigan yuqori nagruzka (high-concurrency) va kirishni nazorat qilish muammolarini hal etish uchun mo'ljallangan. Tizim kriptografik raqamli imzolar yordamida **oflayn (internet yo'q)** rejimda ham xavfsiz ishlashni, hududlarga kirishni darhol tasdiqlashni, WebSockets orqali real vaqtda kuzatishni va og'ir hisobotlarni generatsiya qilishni ta'minlaydi.

---

## 🛠 Texnologiyalar va uskunalar

* **Dasturlash tili:** Go (Golang 1.25)
* **Ma'lumotlar bazasi:** PostgreSQL (Relyatsion saqlash, dinamik filtratsiyalash)
* **Kesh va In-Memory xotira:** Redis (Rollar va ruxsatlarni kesh hisoblash, rate-limiting, hisoblagichlar)
* **Real-time aloqa:** Gorilla WebSocket (Parallel mijozlarni boshqarish)
* **Xavfsizlik va Auth:** JWT (JSON Web Tokens), HMAC-SHA256 (Oflayn QR kodlar uchun raqamli imzo), bcrypt
* **Hisobot va hujjatlar generatori:** `github.com/xuri/excelize/v2` (Excel generatsiyasi), PDF Badge builder
* **Konteynerizatsiya va Infratuzilma:** Docker (Multi-stage build), Docker Compose
* **Test va Benchmarking:** Go standart unit testlari, Vegeta (HTTP nagruzka testi utilitasi)

---

## 🚀 Asosiy imkoniyatlar va amalga oshirilgan ishlar

Tizim loyihalash va amaliyot davomida Clean Architecture tamoyillariga to'liq amal qilgan holda noldan ishlab chiqildi:

### 1. Autentifikatsiya va dinamik RBAC tizimi
* Moslashuvchan rollarni (`SuperAdmin`, `Guard`, `KitchenManager` va h.k.) qo'llab-quvvatlaydigan dinamik **Role-Based Access Control (RBAC)** tizimi qurildi.
* Asosiy ma'lumotlar bazasiga ortiqcha so'rov yubormaslik uchun foydalanuvchi ruxsatlarini mikrosekundlarda tekshiruvchi **Redis Permission Caching** integratsiya qilindi.

### 2. Oflayn rejimda ishlovchi HMAC QR Xavfsizligi
* Ishtirokchilar beydjiklari uchun **HMAC-SHA256 bilan imzolangan QR kodlar** joriy etildi.
* Skanerlash punktlarida internet bo'lmagan taqdirda ham QR kodni soxtalashtirish va ikkinchi marta noqonuniy kirish urinishlarini aniqlash imkonini beradi.

### 3. Real vaqtdagi Monitoring (WebSockets)
* **Gorilla WebSockets** va Go **goroutine**'laridan foydalanib real vaqtdagi hodisalar xabi qurildi.
* Real vaqt rejimida stadionga kirish ko'rsatkichlari, zona sig'imi va ovqatlanish statistikasi monitoring panellariga parallel ravishda uzatib boriladi.

### 4. Kirish va Ovqatlanish nazorati mantig'i
* Har bir zona uchun **IN/OUT (Kirish/Chiqish)** qat'iy nazorati.
* Kunlik **Ovqatlanish limitini nazorat qilish** (har bir ishtirokchi uchun kuniga 1 ta ovqat) va Redis atomic operatsiyalari orqali poyga holatlarini (race-conditions) oldini olish.

### 5. PDF Beydjik va Excel hisobotlarini eksport qilish
* CR80 standartiga mos keluvchi avtomatik PDF beydjiklar generatsiyasi va ularni ommaviy bosmaga chiqarish moduli.
* Moslashuvchan sarlavhalar, avto-kengayuvchi ustunlar va rad etilgan kirish urinishlarini tahlil qilish uchun dinamik PostgreSQL so'rovlari (`UNION ALL`) bilan boyitilgan Excel (`.xlsx`) eksport tizimi.

---

## ⚡ Unumdorlik va Nagruzka Testi Natijalari

Stadion kirish joylarida bir vaqtning o me'yorida minglab sportchilar QR kod skanerlashini simulyatsiya qilish uchun tizim **Vegeta** vositasi orqali jiddiy yuklama ostida test qilindi.

* **Cheklov talabi:** Yuqori nagruzka ostida kechikish (latency) < 500 ms bo'lishi kerak.
* **Test nagruzkasi:** Skanerlash endpointiga soniyasiga 200 ta parallel POST so'rovi (jami 1,000 ta so'rov).

### 📊 Benchmark natijalari:
| Metrika | Ershilgan natija | TZ Talabi |
| :--- | :--- | :--- |
| **P50 Latency (O'rtacha)** | **1.51 ms** | < 500 ms (330 barobar tezroq) |
| **P99 Latency (Eng sekin 1%)** | **2.34 ms** | < 500 ms |
| **Max Latency (Eng uzoq)** | **10.67 ms** | - |
| **Success Rate (Muvaffaqiyat)** | **100.00% (200 OK)** | Birorta ham yo'qotilgan so'rov yo'q |

> **Arxitektura siri:** Yuqori tezlikka ma'lumotlar bazasiga log yozish kabi og'ir jarayonlarni asinxron Go goroutine'lariga o'tkazish va dastlabki kirish ruxsatlarini to'g'ridan-to'g me'yor Redis xotirasida tekshirish orqali erishildi.

---

## 💻 Loyihani ishga tushirish

### Talablar
* [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
* [Go 1.23+](https://golang.org/) (lokal vaqtda kodni yuritish uchun)

### Docker Compose orqali ishga tushirish

1. **Repository'ni klonlang:**
   ```bash
   git clone [https://github.com/Devesanoff/olympic-sport-backend.git](https://github.com/Devesanoff/olympic-sport-backend.git)
   cd olympic-sport-backend




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
