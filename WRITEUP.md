# Engineering Write-Up

## 1. What did you ask the AI to do, and what did you write or decide yourself?

I used Codex to help plan and implement a small Go URL shortener from the take-home prompt. I first asked it to read the prompt, propose an implementation plan, and call out the important design decisions and tradeoffs before writing code.

The main decisions I steered were keeping the project intentionally small, using Go, Dockerizing the service, using an embedded datastore, choosing bbolt over PostgreSQL for v1, returning the same generated code for duplicate URLs, and keeping analytics minimal with a redirect counter.

## 2. Where did you override, correct, or throw away the AI's output, and why?

I kept the implementation focused on the required service behavior instead of adding a larger product surface. The initial planning space included production-style options like PostgreSQL and detailed analytics, but I chose bbolt and a simple stats endpoint because the exercise emphasizes judgment and maintainability over building a large system.

I also kept the HTTP layer on Go's standard library router. A framework would have been fine, but it would not add much value for three endpoints and would make the project feel heavier than necessary.

## 3. The two or three biggest tradeoffs you made, and the alternatives you considered.

The first tradeoff was bbolt versus PostgreSQL. PostgreSQL is the better choice for multi-instance deployments, richer querying, and operational familiarity, but bbolt keeps local setup simple and still gives durable ACID persistence for a single-service take-home.

The second tradeoff was duplicate URL behavior. I chose idempotency for generated links: shortening the same canonical URL twice returns the existing code. The alternative is creating a new code every time, which is useful for campaign tracking but surprising for a basic shortener.

The third tradeoff was analytics scope. I stored only a redirect count. The alternative was recording click events with timestamps, referrers, user agents, and IP-derived metadata, but that adds privacy and data-model complexity beyond the required exercise.

## 4. What's missing, or what would you do with another day?

With another day, I would add link expiration, basic rate limiting, structured request logging, metrics, and a migration path to PostgreSQL for multi-instance deployments. I would also add more analytics if product requirements justified it, especially time-bucketed click counts and referrer summaries.
