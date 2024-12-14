# Movie Platform Microservices

A robust, distributed movie platform built with a microservices architecture, implementing Domain-Driven Design (DDD) principles and enhanced observability.

---

## Architecture Overview

### Services

1. **Movie Service**

   - Core movie management functionality.
   - Establishes gRPC client connections to the Rating and Metadata services.
   - Implements Domain-Driven Design principles.

2. **Rating Service**

   - Manages movie ratings and reviews.
   - Provides gRPC server functionality for rating operations.
   - Uses independent data storage.

3. **Movie Metadata Service**
   - Handles additional movie information.
   - Provides gRPC server functionality for metadata operations.
   - Utilizes specialized metadata storage.

---

## Technical Stack

- **Language**: Golang
- **Service Discovery**: Consul
- **Tracing**: Jaeger
- **Communication**: gRPC
- **Testing**: Test-Driven Development (TDD) with comprehensive test suites
- **Architecture**: Domain-Driven Design (DDD)
- **Observability**: Distributed tracing and metrics collection
- **Data Storage**: MySQL

---

## Service Communication Flow

The communication between services follows a structured flow, ensuring robust integration and observability:

- **Movie Service** communicates with:
  - **Rating Service**: Using gRPC for retrieving and updating movie ratings.
  - **Movie Metadata Service**: Using gRPC to fetch additional metadata.

- **Service Discovery**:
  - Consul manages service registration and dynamic discovery for:
    - Movie Service
    - Rating Service
    - Movie Metadata Service

- **Tracing**:
  - Jaeger monitors and provides end-to-end tracing for all services.
  - Traces requests across the system for diagnostics and performance monitoring.

---

## Key Features

- **gRPC Communication**: Enables efficient, low-latency communication with strong typing.
- **Service Discovery**: Consul is used for dynamic service registration and discovery.
- **Distributed Tracing**: Jaeger provides end-to-end tracing for monitoring requests across the system.
- **Test-Driven Development (TDD)**: Ensures high code quality and comprehensive test coverage.
- **Domain-Driven Design (DDD)**: Focuses on core domain logic for scalable, maintainable services.

---

## Getting Started

### Prerequisites

- Docker and Docker Compose
- gRPC tools
- Consul and Jaeger setup

### Setup Instructions

1. Clone the repository:

   git clone <https://github.com/Sri2103/movies.git>

   ``` bash
   cd movies
   ```

2. Start the services using Docker Compose:

   docker-compose up --build

   ``` bash
   docker run -d \
   -p 8500:8500 \
   -p 8600:8600/udp \
   -p 8300:8300 \
   -p 8301:8301 \
   -p 8301:8301/udp \
   -p 8302:8302 \
   -p 8302:8302/udp \
   --name=dev-consul \
   -e CONSUL_BIND_INTERFACE=eth0 \
   hashicorp/consul:latest agent -server -ui \
   -node=server-1 -bootstrap-expect=1 \
   -client=0.0.0.0

   ```

   ``` bash

   docker run -d --name jaeger \

   -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
   -p 6831:6831/udp \
   -p 6832:6832/udp \
   -p 5778:5778 \
   -p 16686:16686 \
   -p 4317:4317 \
   -p 4318:4318 \
   -p 14250:14250 \
   -p 14268:14268 \
   -p 14269:14269 \
   -p 9411:9411 \
   jaegertracing/all-in-one:latest

   ```

   ``` bash
   docker run --name movieexample_db -e MYSQL_ROOT_PASSWORD=password -e MYSQL_DATABASE=movieexample -p 3306:3306 -d mysql:latest
   ```

3. Access the system:

   - **Consul UI**: `http://localhost:8500`
   - **Jaeger UI**: `http://localhost:16686`

4. Run tests:

   ``` bash
   go test -v ./...
   ```

---

## Service Descriptions

### Movie Service

Manages movies and acts as the central orchestrator, connecting to the Rating and Metadata services for additional operations.

### Rating Service

Handles user ratings and reviews, storing and exposing rating operations.

### Movie Metadata Service

Provides supplementary information about movies, such as genres, cast, and runtime.

---

## Communication Patterns

- Services use gRPC for communication, ensuring strong typing and high performance.
- Service discovery is managed dynamically via Consul.

---

## Observability

### Tracing with Jaeger

- Access Jaeger UI at `http://localhost:16686` to visualize service traces.
- Use the search feature to trace specific requests and identify bottlenecks.

### Metrics Collection

- Collect and analyze metrics to monitor service performance and reliability.

---

## Development Practices

### TDD (Test-Driven Development)

- Write tests before implementing functionality.
- Maintain high test coverage with unit, integration, and end-to-end tests.

### DDD (Domain-Driven Design)

- Divide the system into bounded contexts.
- Focus on core business logic and domain-specific requirements.

---

## Deployment

### Docker Deployment

1. Build Docker images:
   docker build -t movie-service .
   docker build -t rating-service .
   docker build -t metadata-service .
2. Deploy services using Docker Compose:
   docker-compose up -d
3. Access the system:
   - Consul UI: <http://localhost:8500>
   - Jaeger UI: <http://localhost:16686>

### Kubernetes Deployment

1. Configure Kubernetes cluster and apply the deployment files.
  Use Helm to manage Kubernetes deployments.
  Repo: <https://github.com/Sri2103/moviesDeployment>
  ToStart:
  
   movie:

   ```bash
      helm install movie movie/
   ```

   Tracing:

   ```bash
      helm install bitnami/jaeger
      ```

   Consul:

   ```bash
      helm install consul hashicorp/consul
   ```

      rating:

   ```bash
      helm install rating rating/
    ```

   metatadata:

   ```bash
       helm install metatadata metatadata/ 
   ```

## Contributing

1. Fork the repository.
2. Create a feature branch:

   git checkout -b feature/your-feature

3. Commit your changes and push them to the branch:

   git push origin feature/your-feature

4. Submit a pull request for review.

---
