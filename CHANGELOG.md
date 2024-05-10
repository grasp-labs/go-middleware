# Changelog

## v1.1.1 - 2024-05-10

### Fixes

* Permission filter: add url as optional field for better testing purposes

## v1.1.0 - 2024-05-09

### Fixes

* Custom context: fix broken test.
* Log: change error type on invalid context to return echo.NewHTTPError.
* Usage: change error type on invalid context to return echo.NewHTTPError.

### Enhancements

* Permission filter: add permission filter middleware.

### Docs

* CHANGELOG.md: add changelog file.

## v1.0.3 - 2024-04-24

### Fixes

* Custom context: change header request id name from request-id to X-Request-Id.

## v1.0.2 - 2024-04-23

### Fixes

* Usage: fix typo for building mode from tests to test.

## v1.0.1 - 2024-04-23

### Enhancements

* Create middleware for: JWT, custom context, log and usage.

## v1.0.0 - 2024-04-23

### Enhancements

* Create repository
