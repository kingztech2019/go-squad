# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-06-11

### Added
- `TransactionService` — payment initiation, verification, refunds, USSD bank list, missed webhooks
- `VirtualAccountService` — NUBAN virtual account creation, queries, updates, sandbox simulation
- `TransferService` — fund transfers, intra-Squad transfers, account lookup, transfer history
- `DisputeService` — dispute listing, detail, evidence upload (multipart), accept/reject
- `VASService` — airtime, data bundles, cable TV, electricity, SMS
- `ParseWebhook` and `VerifySignature` — HMAC-SHA512 webhook validation
- Typed webhook event bodies for all Squad event types
- Functional options: `WithSandbox`, `WithProduction`, `WithBaseURL`, `WithHTTPClient`, `WithTimeout`, `WithUserAgent`
- Auto-detection of sandbox environment from `sandbox_sk_` key prefix
- Zero external runtime dependencies
- Full test suite using `net/http/httptest` mocks
- Examples: payment flow, virtual accounts, webhook server
- GitHub Actions CI matrix across Go 1.21, 1.22, 1.23
