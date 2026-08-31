# Business Engine

A transactional backend for commerce operations involving **orders, products, inventory, scheduling, and payments**.

Business Engine is a backend project built with Go and PostgreSQL that models the core operations of a business where a single order can combine physical products and scheduled services.

The main focus of the project is not CRUD itself, but the **business rules, transactional consistency, concurrency control, and coordination between different domains** involved in processing an order.

## Overview

A customer can create an order containing different types of items:

* **Products**, which are fulfilled through the Inventory domain.
* **Scheduled services**, which are fulfilled through the Scheduling domain.

An order remains `PENDING` while it is being built. Inventory and scheduling operations are coordinated when the order is processed, allowing the system to preserve its business invariants even when concurrent requests target the same resources.

The system is currently organized around four main domains:

```text
                    ┌──────────────────────┐
                    │       Sales          │
                    │       Orders         │
                    └──────────┬───────────┘
                               │
                    ┌──────────┼──────────┐
                    │          │          │
                    ▼          ▼          ▼
               Inventory  Scheduling   Payment
                    │          │          │
                    └──────────┼──────────┘
                               │
                         PostgreSQL
```

## Domains

### Sales

Responsible for the order lifecycle.

An order can contain both products and scheduled services. Products and services are represented as order items while the order is in a `PENDING` state.

Main operations include:

* Create order
* Add product
* Add appointment
* Process order
* Cancel order
* Create return
* Retrieve order

### Inventory

Responsible for product stock and its invariants.

The inventory domain supports:

* Stock queries
* Purchases
* Sales
* Stock adjustments
* Product activation/deactivation

Stock mutations are designed to be safe under concurrent operations.

### Scheduling

Responsible for resources, availability and appointments.

The scheduling domain supports:

* Resource management
* Availability periods
* Appointment creation
* Rescheduling
* Cancellation
* Availability validation

PostgreSQL range types and exclusion constraints are used to prevent overlapping availability periods.

### Payment

Responsible for the payment lifecycle associated with an order/appointment.

Payments have an explicit state machine and enforce rules such as:

* Only valid payment state transitions are allowed.
* A completed appointment cannot have multiple completed payments.
* Idempotency keys are unique.
* Payment amounts must be positive.
* Currency and status values are constrained at the database level.

## Architecture

The repository is organized as a multi-module Go project:

```text
business-engine/
│
├── business-engine-service/
│   ├── cmd/
│   ├── dto/
│   ├── services/
│   └── tests/
│
├── postgres/
│   ├── config/
│   ├── migration/
│   ├── model/
│   └── services/
│
├── infra/
│   └── ...
│
└── Makefile
```

### `business-engine-service`

The main backend service.

This module contains the application contracts, DTOs, and service interfaces that define the business operations exposed by the application.

```text
github.com/tcero76/business-engine/business-engine-service
```

### `postgres`

The PostgreSQL implementation layer.

It contains the GORM models, database configuration, migrations and concrete implementations of the service interfaces.

```text
github.com/tcero76/business-engine/postgres
```

This separation allows the business service to define its contracts while the PostgreSQL module provides the persistence implementation.

## Concurrency and Transactions

One of the main goals of the project is handling **concurrent business operations correctly**.

Examples include:

* Two users attempting to purchase the last available units of a product.
* Two users attempting to book the same resource at the same time.
* Concurrent stock adjustments.
* Concurrent appointment creation and rescheduling.

The implementation relies on PostgreSQL transactions, row-level locking, atomic updates and database constraints where appropriate.

The database is treated as an active part of the domain model rather than merely as a persistence mechanism.

## Database

PostgreSQL is used as the transactional database.

The schema is divided into domains:

```text
demo
├── users

sales
├── orders
└── order_items

inventory
├── products
└── stock

scheduling
├── resources
├── availability
└── appointments

payment
└── payments
```

Important business invariants are enforced at the database level whenever possible through:

* Foreign keys
* Unique constraints
* Check constraints
* Partial unique indexes
* PostgreSQL range types
* Exclusion constraints
* Transactional locking

## Testing

The project includes integration tests covering the main business domains.

Particular attention is given to concurrent operations and scenarios where multiple transactions attempt to modify the same business resource.

The intention is to test not only the happy path, but also the invariants that must hold when operations race against each other.

## Running the project

### Requirements

* Go
* Docker
* Docker Compose
* Make

The development environment uses Docker Compose for PostgreSQL and the application infrastructure.

Start the environment with:

```bash
make up
```

Check the running services with:

```bash
make ps
```

Stop the environment with:

```bash
make down
```
## Development

The Go service uses [Air](https://github.com/air-verse/air) for live reloading and [Delve](https://github.com/go-delve/delve) for debugging during development.

The service is built with:

* Go
* GORM
* PostgreSQL
* Docker
* Docker Compose
* Delve
* Air

## Design Goals

The project is intended as an exploration of backend engineering problems that appear once a system moves beyond simple CRUD operations.

The main goals are:

* Model business rules explicitly.
* Preserve domain invariants under concurrent access.
* Use database transactions deliberately.
* Keep service contracts separate from their persistence implementations.
* Prefer database constraints for invariants that belong to the data model.
* Make failure scenarios explicit.
* Build a foundation that can evolve toward asynchronous processing.

## Roadmap

The project is designed to evolve toward an event-driven architecture.

Planned components include:

* [ ] Transactional Outbox Pattern
* [ ] Message broker integration
* [ ] Asynchronous workers
* [ ] Electronic document generation and integration with the Chilean SII
* [ ] Customer notification service
* [ ] Additional observability
* [ ] API documentation

The long-term goal is to use the transactional core as the source of reliable business events that can be consumed by independent services.

## Why this project?

Business Engine is primarily a backend engineering project.

Rather than focusing on building a large number of endpoints, the project explores the problems that arise when multiple business domains interact:

> What happens when two customers try to buy the last product at the same time?

> What happens when two customers try to book the same resource simultaneously?

> What happens when processing an order requires changes across multiple domains?

> Which invariants belong in application code, and which should be guaranteed by the database?

These questions drive the architecture and implementation decisions throughout the project.
