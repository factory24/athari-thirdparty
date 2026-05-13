# Athari ThirdParty Wrappers

`athari-thirdparty` is a collection of centralized wrappers and clients for third-party services used within the Flow ecosystem. It simplifies integration with external platforms by providing idiomatic Go interfaces and standardizing configuration.

## Installation

To install the package:

```bash
go get github.com/factory24/athari-thirdparty
```

## Available Wrappers

### 1. Azure (`azure/`)
Provides clients for Azure services:
- **Service Bus**: Simplified message sending and receiving.
- **Direct Method**: Interface for Azure IoT Hub direct methods.

### 2. Apache Pulsar (`pulsar/`)
A high-level wrapper for the Pulsar Go client, supporting:
- Multi-topic producers and consumers.
- Automatic reconnection and standardized logging.
- JSON-based event processing using `flow-system` event structures.

### 3. Sentry (`sentry/`)
Streamlined Sentry initialization and error reporting.

**Usage:**
```go
import "github.com/factory24/athari-thirdparty/sentry"

sentryClient := sentryClient.NewSentryClient(cfg.GetSentryConfig())
sentryClient.Connect()
defer sentryClient.Flush(2 * time.Second)
```

### 4. Elasticsearch (`elasticsearch/`)
Wrapper for the Elasticsearch v8 client for indexing and searching documents.

### 5. Infisical (`infisical/`)
Integration with Infisical for secret management and environment variable injection.

### 6. ChirpStack (`chirpstack/`)
Clients for interacting with the ChirpStack LoRaWAN Network Server API.

### 7. Payments (`payments/`)
Standardized interfaces for multiple payment gateways (e.g., Arkesel, Hubtel).

## Dependencies

This package depends on `github.com/factory24/flow-system` for shared configurations and models.

## Usage Guidelines

- Always use the provided factory functions (e.g., `New...Client`) to instantiate wrappers.
- Ensure that credentials and connection strings are managed via `flow-system/pkg/config`.
