## Technical Brief Template

This template helps you write clear, actionable briefs for AI-powered development tasks.

---

## 1. Title of the task

Write here:

`<Descriptive title of the functionality or service>`

Describe **what** will be built in a concise and explicit way. Prefer concrete, user-facing language over generic names.

- **Bad**: Recommendation feature  
- **Good**: Outfit recommender for cold-weather workdays using purchase history and style quiz

---

## 2. Context

Give details about the current behavior, why it is a problem and what's the expected goal with this task.

Include:

- **Current behavior**: How the app works today
- **Pain points**: What is going wrong
- **Execution context**: Which modules or services are involved
- **Objective**: What's the goal with this task

Example:

- The current “You may also like” section shows products only based on global popularity.  
- This leads to irrelevant suggestions (summer dresses during winter, party outfits for users browsing office looks).  
- The new service will run on the product detail page and the home feed.  
- The objective is to generate weather and occasion-aware outfit suggestions to increase PDP add‑to‑cart by 8% and reduce return rate related to “not suitable for the occasion”.

---

## 3. Technical requirements

Describe **how** you want this solution to be implemented.

### 3.1 Language / Stack

Specify the exact technologies and versions.

- **Language**:  
- **Minimum version**:  
- **Framework(s)**:  
- **Database / Storage**:  
- **AI / ML stack** (if relevant): model provider, endpoints, model names, vector DB, etc.

Fashion app example:

- **Language**: Go  
- **Minimum version**: Go 1.22+  
- **HTTP / routing**: `net/http` with a thin router (e.g. Gin, Echo)  
- **Clients**: Mobile app calls this service 
- **Database**: PostgreSQL (e.g. `database/sql` + `pgx` or sqlc)  
- **AI**: External LLM call to Gemini 3.1 Pro for natural language styling advice

### 3.2 Architecture and Patterns

Clarify **architecture principles** and **patterns** the solution must follow.

Consider:

- Domain boundaries (e.g., `user-profile`, `catalog`, `styling`, `recommendations`)  
- Service type (microservice, module inside monolith, background worker, etc.)  
- Patterns (Clean Architecture, hexagonal, etc.)  
- Communication (REST, GraphQL, gRPC, events)

Fashion app example:

- Follow Clean Architecture with clear separation between:  
  - **Domain**: outfit recommendation rules and constraints (weather, dress code, user preferences)  
  - **Application**: use cases: `GenerateOutfitSuggestions`, `RankOutfitsForPDP`  
  - **Infrastructure**: LLM client, catalog repository, weather API client  
- Use Strategy Pattern for outfit ranking (different strategies for “work”, “casual weekend”, “formal event”).  
- Expose a REST endpoint: `POST /styling/outfit-suggestions`.  

### 3.3 Inputs

Define **exactly** what the service/function receives.

Describe:

- Field names  
- Types  
- Optional vs required  
- Example values

Example (Go structs):

```go
type Occasion string

const (
	OccasionWork    Occasion = "work"
	OccasionDate    Occasion = "date"
	OccasionWedding Occasion = "wedding"
	OccasionParty   Occasion = "party"
	OccasionCasual  Occasion = "casual"
)

type BudgetRange struct {
	Min float64
	Max float64
}

// GenerateOutfitInput is the request body for POST /styling/outfit-suggestions.
type GenerateOutfitInput struct {
	UserID        string    // required
	Occasion      Occasion  // required
	Location      string    // city or geo code for weather
	EventDate     *string   // optional: ISO date for seasonal / weather context
	BudgetRange   *BudgetRange
}
```

### 3.4 Outputs

Define the **response shape** and expectations.

Include:

- Fields and types  
- How many items  
- Order guarantees (sorted by score, price, relevance, etc.)  
- Error handling format

Example (Go structs):

```go
type ItemCategory string

const (
	CategoryTop       ItemCategory = "top"
	CategoryBottom    ItemCategory = "bottom"
	CategoryOuterwear ItemCategory = "outerwear"
	CategoryShoes     ItemCategory = "shoes"
	CategoryAccessory ItemCategory = "accessory"
)

type OutfitItem struct {
	ProductID string
	Category  ItemCategory
	Reason    string // e.g. "matches your minimal style and budget"
}

type Outfit struct {
	ID         string
	Items      []OutfitItem
	TotalPrice float64
	Score      float64
}

type GenerateOutfitResponse struct {
	Outfits []Outfit // always return 3 results
	TraceID string
}
```

---

## 4. Constraints

List **what the AI MUST NOT do** and any strong preferences or limits.

Consider:

- **Libraries**: which are forbidden  
- **Performance**: throughput, timeouts, batch sizes  
- **Privacy / Compliance**: no storing PII beyond X, anonymization rules 
- **Domain limitations**: what the stylist must avoid (e.g., no suggestions in restricted categories, no speculation about medical or body-image topics)  
- **UX guardrails**: tone of voice, languages, how explanations are phrased

Fashion app examples:

- Do **not** call any unreleased or experimental AI endpoints; only use `styling-llm-v2` via the `llmClient` wrapper.  
- Do **not** block the main product page longer than 300 ms on this service; if AI is slow, return cached outfits or a graceful fallback.  
- Do **not** store raw user measurements in logs; only anonymized size buckets.  
- Avoid generating suggestions that:  
  - reference sensitive body-image topics,  
  - assume gender based only on name,  
  - include out-of-stock products.  
- Maintain a friendly, encouraging tone in user-facing explanations (e.g., “This blazer keeps your look polished and warm”).

---

## 5. Definition of Done (DoD)

Specify **objective criteria** that must be true for the work to be considered finished.

Include:

- **Code quality**: linting, formatting, typing  
- **Testing**: unit, integration, coverage thresholds, mock behavior for AI calls  

Fashion app example:

- **Code & Quality**
  - The module follows the agreed architecture (e.g., Clean Architecture, domain separated from infrastructure).  
  - Idiomatic Go: explicit errors, no unnecessary `interface{}` / `any` in domain code; table-driven tests where it helps.  
  - Code passes `gofmt` / `go vet` and the team’s static checks (e.g. `golangci-lint`, `staticcheck`).

- **Testing**
  - At least **85% unit test coverage** for the new domain and application layer code.  
  - Unit tests cover:  
    - outfit ranking logic (different occasions, budgets, and style tags),  
    - behavior when weather API fails,  
    - fallbacks when AI response is missing or malformed.  
  - Integration tests for:  
    - `POST /styling/outfit-suggestions` happy path with mocked AI and catalog,  
    - error responses for invalid input (missing `userId`, unsupported `occasion`, etc.).
