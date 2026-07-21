# Auth Specification

## Purpose

User registration, login, JWT access/refresh token issuance, and logout for individual users. All protected endpoints require a valid JWT.

## Requirements

### Requirement: User Registration

The system MUST accept email and password, validate uniqueness of email, hash the password, and return the created user.

#### Scenario: Successful registration

- GIVEN a valid email and password (min 8 chars) not already in the system
- WHEN the user submits POST /auth/register
- THEN the system creates a user record with hashed password
- AND returns 201 with user ID and email

#### Scenario: Duplicate email

- GIVEN an email already registered
- WHEN the user submits POST /auth/register
- THEN the system returns 409 with an error message

#### Scenario: Invalid input

- GIVEN a malformed email or password shorter than 8 characters
- WHEN the user submits POST /auth/register
- THEN the system returns 400 with validation errors

### Requirement: User Login

The system MUST validate credentials and issue an access token (short-lived) and a refresh token (long-lived).

#### Scenario: Successful login

- GIVEN a registered user with correct credentials
- WHEN the user submits POST /auth/login
- THEN the system returns 200 with access_token and refresh_token

#### Scenario: Invalid credentials

- GIVEN an email that exists but wrong password
- WHEN the user submits POST /auth/login
- THEN the system returns 401 (MUST NOT reveal whether email exists)

### Requirement: Token Refresh

The system MUST accept a valid refresh token and issue a new access token. The old refresh token MUST be rotated (invalidated).

#### Scenario: Successful refresh

- GIVEN a valid, non-expired refresh token
- WHEN the user submits POST /auth/refresh
- THEN the system returns 200 with a new access_token and refresh_token
- AND the old refresh token is invalidated

#### Scenario: Expired refresh token

- GIVEN an expired or already-used refresh token
- WHEN the user submits POST /auth/refresh
- THEN the system returns 401

### Requirement: Logout

The system MUST invalidate the user's refresh token(s).

#### Scenario: Successful logout

- GIVEN an authenticated user with a valid refresh token
- WHEN the user submits POST /auth/logout
- THEN the system invalidates the refresh token
- AND returns 200

### Requirement: JWT Auth Middleware

The system MUST protect all non-auth endpoints by validating the JWT access token from the Authorization header.

#### Scenario: Valid token

- GIVEN a request with a valid, non-expired JWT in the Authorization header
- WHEN the request reaches a protected endpoint
- THEN the system extracts user_id from the token and proceeds

#### Scenario: Missing or invalid token

- GIVEN a request without a token or with an expired/malformed token
- WHEN the request reaches a protected endpoint
- THEN the system returns 401

## Constraints

- Passwords MUST be hashed with bcrypt (cost >= 10)
- Access tokens MUST expire in 15 minutes; refresh tokens in 7 days
- All auth endpoints MUST be rate-limited (future consideration)
- Tokens MUST NOT contain sensitive data beyond user_id and expiration

## Dependencies

- None (foundational capability)
