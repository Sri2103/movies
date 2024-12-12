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

- **Service Discovery**: Consul
- **Tracing**: Jaeger
- **Communication**: gRPC
- **Testing**: Test-Driven Development (TDD) with comprehensive test suites
- **Architecture**: Domain-Driven Design (DDD)
- **Observability**: Distributed tracing and metrics collection

---

mermaid
graph LR
    A[Movie Service] -->|gRPC| B[Rating Service]
    A -->|gRPC| C[Movie Metadata Service]
    D[Consul] -->|Service Discovery| A
    D -->|Service Discovery| B
    D -->|Service Discovery| C
    E[Jaeger] -->|Tracing| A
    E -->|Tracing| B
    E -->|Tracing| C


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
   ```bash
   git clone https://github.com/your-repo/movie-platform-microservices.git
   cd movie-platform-microservices
````

2. Start the services using Docker Compose:

   ```bash
   docker-compose up --build
   ```

3. Access the system:

   - **Consul UI**: `http://localhost:8500`
   - **Jaeger UI**: `http://localhost:16686`

4. Run tests:
   ```bash
   npm test
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

## Contributing

1. Fork the repository.
2. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature
   ```
3. Commit your changes and push them to the branch:
   ```bash
   git push origin feature/your-feature
   ```
4. Submit a pull request for review.

---

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
