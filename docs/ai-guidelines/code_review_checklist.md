# Code review checklist (human review of AI-generated code)

Use this when reviewing changes produced by an AI so you can catch hallucinations, logic mistakes, security gaps, and performance traps. Examples assume a **Go** backend for a **clothing recommendation** app (catalog, outfits, style quiz, weather-aware suggestions).

---

## 1. Hallucinations (packages, APIs, and references)

### Key question

Does that import, symbol, or API actually exist?

Models can invent:

- Go module paths or package names
- Functions, methods, and struct fields
- HTTP routes, query parameters, or third-party SDK calls
- “Standard” patterns that do not match real `go doc` behavior

### Checklist

- [ ] `go build` / `go test` pass on the intended Go version (covers bogus imports, missing symbols, and stdlib APIs not available for that version)
- [ ] HTTP client calls match the real base URL, paths, and auth headers of your catalog or ML service
- [ ] Database drivers and SQL dialect match what you run in prod (e.g. Postgres vs SQLite-only syntax)

---

## 2. Business logic and alignment with the brief

### Key question

Does the behavior match the product rules and the brief, not only “it compiles”?

**Common AI mistakes:**

- Wrong scoring or ranking (popularity vs personalization vs weather)
- Incorrect handling of sizes, inventory, or “in stock” flags
- Time and timezone bugs for “seasonal” or event-based outfits
- Using `float64` for prices or discounts
- Ignoring constraints from the brief (e.g. only office-appropriate items, exclude certain categories)

### Checklist

- [ ] Recommendation rules match the brief (e.g. cold-weather workday vs party context)
- [ ] Edge cases are handled: empty closet, new user, missing quiz answers, out-of-stock SKUs
- [ ] Money uses integer minor units or `decimal`-style types, not raw floats
- [ ] Dates and “season” logic are consistent (timezone, locale)
- [ ] The implementation does not contradict explicit non-goals or exclusions in the brief
- [ ] Definition of Done in the brief is satisfied (API shape, events, metrics, etc.)

---

## 3. Security

### Key question

Is the code safe to run with real users and real catalog data?

AI often produces code that “works” but widens the attack surface. The checklist below is the concrete pass; use it instead of re-listing the same themes in prose.

### Checklist

- [ ] Inputs are validated and bounded (pagination limits, max items per outfit, allowed sort fields)
- [ ] SQL uses parameterized queries; no string-concatenated `WHERE` from user input
- [ ] No command execution with unsanitized user strings (e.g. shelling out for “image processing”)
- [ ] Secrets come from env or a secret manager, not source code
- [ ] Authorization checks that a user can only access their own closet or recommendations
- [ ] Logs and traces do not include tokens, full payment payloads, or full profile blobs

---

## 4. Performance and scalability

### Key question

Will this stay fast and stable as catalog size and traffic grow?

**Typical issues:** unbounded reads, N+1 queries, missing pagination, goroutine/channel leaks on external calls, oversized JSON.

### Checklist

- [ ] List and search endpoints use pagination (cursor or offset/limit with sane caps)
- [ ] Hot paths avoid loading full catalogs; use indexes, limits, and batch queries
- [ ] No N+1 patterns when joining products, images, and inventory for a recommendation set
- [ ] External calls (ML scoring, weather API) have timeouts, retries with backoff, and circuit breaking where appropriate
- [ ] Caching is considered for expensive reads (e.g. popular outfits, static taxonomy) without stale incorrect inventory
- [ ] Memory: streaming or chunking for large exports; bounded buffers for embedding or image pipelines

---

## Review outcome

Roll-up of sections above (same bar, one place to sign off):

- [ ] Section 1 — No hallucinated or mismatched APIs/packages
- [ ] Section 2 — Logic and brief alignment verified
- [ ] Section 3 — Security reviewed
- [ ] Section 4 — Performance and scalability reviewed
- [ ] Necessary fixes are done or tracked
- [ ] Ready to merge or commit