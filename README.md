# Flarecloud

This project aims to develop a robust, secure and scalable Content Delivery Network (CDN) in Go, integrating features such as HTTP proxy, caching, load balancing and advanced security.

## Description

The goal is to design a CDN server capable of:

- **Efficiently routing HTTP requests** via a proxy.
- **Optimizing latency** through in-memory caching (LRU) and optionally distributed caching (Redis).
- **Load balancing** across multiple servers using different algorithms (Round Robin, Weighted RR, Least Connections).
- **Securing traffic** by providing HTTPS via Let's Encrypt, TLS 1.3, DDoS attack protection and rate limiting implementation.
- **Monitoring the service** in real-time with Prometheus, Grafana and Loki for log management.

## Technologies Used

- **Main Language:** Go
- **Modules & Packages:**
  - `net/http` for HTTP request handling.
  - `hashicorp/golang-lru` for in-memory cache implementation.
  - `golang.org/x/time/rate` for rate limiting management.
- **Security:**
  - Let's Encrypt and TLS 1.3 for HTTPS.
  - DDoS attack prevention and rate limiting mechanisms.
- **Deployment:**
  - Docker and Kubernetes for containerization and orchestration.
- **Monitoring:**
  - Prometheus and Grafana for performance tracking.
  - Loki for log aggregation.

## Functional Objectives

- **HTTP Proxy Server:** Intelligently direct requests to cache or origin servers.
- **Efficient Cache:** Minimize response times by temporarily storing content.
- **Load Balancer:** Distribute request load for better scalability.
- **Advanced Security:** Ensure secure connections via HTTPS and anti-DDoS techniques.
- **CI/CD & Testing:** Automated pipeline for deployment and continuous testing (unit and integration).

## Installation & Deployment

1. **Clone the Repository:**

   ```bash
   git clone https://github.com/Triips-TheCoder/Flarecloud.git
   cd Flarecloud
   ```

2. **Install Dependencies:**

   - Install [Go](https://golang.org/dl/) (recent version).
   - Install Docker and Kubernetes (Minikube for local).

3. **Start the Server:**

   ```bash
   go run main.go
   ```

4. **Run Tests:**
   ```bash
   go test ./...
   ```


## Authors & Acknowledgments

- **Triips-TheCoder** - *Developer* - [GitHub](https://github.com/Triips-TheCoder)
- **NicoooM** - *Developer* - [GitHub](https://github.com/NicoooM)
- **lucasboucher** - *Developer* - [GitHub](https://github.com/lucasboucher)
- **PaulMazeau** - *Developer* - [GitHub](https://github.com/PaulMazeau)