# Time Travel Record API
## Prathamesh Kulkarni (prathamesh.kulkarni@okstate.edu)

A versioned, auditable record-keeping system built in Go to simulate a simplified insurance platform backend.

This project enhances a basic REST API by introducing:

- Persistent storage using SQLite
- Record versioning and history tracking
- Time-travel queries for historical reconstruction
- Full backward compatibility with v1 endpoints

---

## Problem Overview

In insurance systems, storing only the current state of data is not sufficient.

Policyholders may update risk-related information (business hours, workforce size, liability limits). If these changes are reported late, insurers must reconstruct:

- What data was known
- When it was known
- When the change actually occurred

This system provides a historical, versioned view of records to support:

- Compliance audits
- Retroactive premium adjustments
- Risk reassessments
- Regulatory reporting

---

# Objectives Completed

## Objective 1 – Persistent Storage

Replaced in-memory storage with SQLite to ensure:

- Data durability across server restarts
- Structured relational schema
- Reliable storage for compliance

## Objective 2 – Time Travel Functionality

Implemented record versioning with:

- Immutable historical versions
- Version-based retrieval
- Version listing per record
- Backward compatibility with `/api/v1`

---

# Running the Application

## Clone the Repository

```bash
git clone https://github.com/<your-username>/timetravel-api.git
cd timetravel
```
## Run the Server
```bash
go run .
```

## Health Check
```bash
curl -X POST http://localhost:8000/api/v1/health
```
## Expected Response
```bash
{"ok": true}
```

## API Reference
## V1 (Backward Compatible)
```bash
GET /api/v1/records/{id}
POST /api/v1/records/{id}
```
## V2 (Versioned + Time Travel)
### Retrieve Latest Version
```bash
GET /api/v2/records/{id}
```
### Retrieve Specific Version
```bash
GET /api/v2/records/{id}?version=3
```
### List All Versions
```bash
GET /api/v2/records/{id}/versions
```
### Update Record (Creates New Version)
```bash
POST /api/v2/records/{id}
```
## Database Design
SQLite schema includes:

1. records table (record metadata)

2. record_versions table (versioned snapshots)

Each update results in:

-Incremented version number

-Timestamped snapshot of record state

-Immutable historical storage

## Future Scope
Delta-based storage optimization

Effective-date tracking vs system timestamp

Soft-delete support

User attribution per change

Migration to PostgreSQL for production scale



