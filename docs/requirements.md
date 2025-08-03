# go-auth requirements
This document outlines a set of requirements for the go-auth service.

## functional
### GAR-F-1 -- authentication
The service must support user authentication via a username and password combination.

### GAR-F-2 -- two factor authentication
Users must have the option to enable Time-Based One-Time Password (TOTP) as a second factor for authentication.

### GAR-F-3 -- the identity management cycle
Service administrators must be able to perform identity lifecycle operations, including provisioning, management, and deprovisioning of user accounts.

### GAR-F-4 -- single sign on
Upon successful authentication, the service must issue a secure access JWT token that can be used to access other services within the ecosystem.

### GAR-F-5 -- role-based authorization
The system must support a Role-Based Access Control (RBAC) model for access management. Administrators shall assign one or more roles to an identity. A role is a collection of granular permissions.
Access to a protected resource (e.g., an API endpoint) must be enforced by checking for the presence of a specific, required permission within the user's identity token. The token must contain an aggregated list of all permissions granted to the user through their assigned roles.

### GAR-F-6 -- interface
The service should be available as a grpc and rest interface.

### GAR-F-7 -- oauth2.0 compliance
The rest interface should be OAuth 2.0 compliant.

### GAR-F-8 -- Password Management Policies
Service administrators must be able to configure password policies, including minimum length, complexity requirements (e.g., uppercase, numbers, symbols), and expiration period. The system should also prevent reuse of recent passwords.

### GAR-F-9 -- Secure Password Reset
The service must provide a secure flow for users to reset their forgotten passwords (e.g., via an email link with a short-lived token).

### GAR-F-10 -- Brute-force Attack Protection
The service must implement mechanisms to mitigate brute-force attacks, such as rate limiting on login attempts and temporary account lockout after a certain number of failed attempts.

### GAR-F-11 -- Audit Logging
All security-sensitive events must be logged for audit purposes. This includes login attempts (successful and failed), password changes, role modifications, and token issuance/revocation. Logs must contain information about who performed the action, what action was performed, the target resource, and the timestamp.

### GAR-F-12 -- Token Revocation
The service must provide an endpoint for revoking active access tokens (e.g., upon user logout or if a token is compromised).

### GAR-F-13 -- Service Accounts / M2M Authentication
The service must support authentication for non-human clients (service accounts) using a client ID and client secret (implementing the OAuth 2.0 Client Credentials Flow).

## non-functional
### GAR-NF-1 -- the directory
All the identities, as well as the permissions and the roles, and all the relations should be stored inside a single database (might be sharded).

### GAR-NF-2 -- qos
The authentication and the authorization requests should be answered by the service in less than 151ms (p95). \
The management cycle requests should be answered in less than 501ms (p95).

### GAR-NF-3 -- caching and invalidation
To increase the service speed and resilience performance, the authentication and authorization requests might be cached. The cache invalidation strategy must ensure data consistency. For example, any change to a user's roles or permissions must immediately invalidate relevant cached entries.

### GAR-NF-4 -- test coverage
The service should have a >=80% test coverage for the api. \
Unit test coverage is advised for the critical code sections.

### GAR-NF-5 -- the directory robustness
The directory service must be highly available and fault-tolerant, utilizing an automatic failover mechanism (e.g., a replica set).

### GAR-NF-6 -- Security: Data at Rest and in Transit
All network communication must be encrypted using TLS 1.2 or higher. Sensitive data stored in the database, especially user credentials, must be securely stored.

### GAR-NF-7 -- Observability: Metrics and Logging
The service must expose key performance and health metrics in a standard format (e.g., Prometheus). It should produce structured logs (e.g., JSON) to facilitate automated parsing and analysis.

### GAR-NF-8 -- Scalability
The service architecture should be horizontally scalable. It must be able to handle a load of 1000 requests per second for authentication/authorization while meeting the QoS requirements (GAR-NF-2).

### GAR-NF-9 -- Documentation
The REST API must be documented using the OpenAPI (Swagger) specification. 
The GRPC protobuf must be available as a set of .proto files.
