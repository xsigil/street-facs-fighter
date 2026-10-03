# Contributing to Street FACS Fighter 🥊

First off, thank you for considering contributing to **Street FACS Fighter**!

Whether you are a **Go software engineer**, a **certified FACS coder**, an **NLP practitioner**, or a **cognitive/behavioral science researcher**, your insights and contributions are warmly welcome.

---

## 🧭 Project Vision

Street FACS Fighter merges **hardcore enterprise software architecture (DDD in Go)** with **retro terminal gaming (Sixel rasterization & retro waveform synthesis)** to build the ultimate training ground for real-time nonverbal calibration.

We especially welcome contributions at the intersection of:
1. **FACS Precision**: Incorporating formal scoring rules (e.g., intensity, asymmetry, timing).
2. **Terminal UX**: Improving sixel support, cross-platform terminal raw mode, and latency.
3. **Open Science**: Facilitating connections with open academic facial expression databases.

---

## 🔬 Ideas for FACS Specialists & Behavioral Researchers

We are actively seeking domain expertise! Here are some features and enhancements we would love your help with:

### 1. Intensity Scoring (A–E) Multipliers
* **Context**: FACS codes intensity from `A` (trace) through `E` (extreme).
* **Idea**: Award critical damage multipliers for exact intensity matches (e.g., hitting `12c` instead of just `12`), or deliver partial damage for close estimations.

### 2. Micro-Expression Flash Mode (METT / Sub-200ms)
* **Context**: Real-world micro-expressions flash between 40ms and 200ms before returning to baseline.
* **Idea**: Implement a high-speed training mode where the Sixel image flashes for 100–250ms and is immediately masked, testing pure sensory acuity and retention.

### 3. Asymmetry & Unilateral Movements (L / R)
* **Context**: Facial expressions of contempt or insincerity often exhibit unilateral activation (e.g., `R14`, `L12`).
* **Idea**: Extend entity modeling and input parsing to distinguish left/right facial activations.

### 4. Academic Open Dataset Ingest Adapters
* **Context**: To protect copyrights, proprietary test booklets are excluded. However, open research databases exist (such as CK+, DISFA, or RAF-DB).
* **Idea**: Build CLI importers under `cmd/` to parse standard CSV/annotation formats from open datasets directly into `app.db`.

---

## 🏛️ Architectural Guidelines (DDD)

We follow strict **Domain-Driven Design (DDD)** principles. When adding features, please respect layer boundaries:

```
internal/
├── domain/            # Pure business entities & interfaces (Zero external dependencies)
│   ├── entity/        # Question, TargetAU, Player, CombatSession
│   └── repository/    # Interface definitions (e.g., IQuestionRepository)
├── usecase/           # Application business logic (Combat progression, scoring rules)
└── infrastructure/    # External drivers & I/O
    ├── sqlite3/       # Database implementations & migrations
    └── sixel/         # Terminal renderers and terminal control
```

* **Domain entities** must remain pure and free from terminal/OS/SQL dependencies.
* **Usecases** should encapsulate combat progression and scoring mechanics.
* **Infrastructure** handles Sixel encoding, audio subprocesses, and SQLite drivers.

---

## 🛠️ Development Workflow

### Prerequisites
* Go 1.20+
* A Sixel-capable terminal emulator (WezTerm recommended)
* `sqlite3`

### Local Setup
```bash
# Fork & clone
git clone https://github.com/<your-username>/street-facs-fighter.git
cd street-facs-fighter

# Install Go modules
go mod tidy

# Generate local 8-bit sound effects
go run ./cmd/gen-se

# Run tests
go test ./...
```

### Pull Request Process

1. **Create a topic branch**:
   ```bash
   git checkout -b feature/intensity-scoring
   ```
2. **Follow Go conventions**:
   Ensure your code is properly formatted (`go fmt ./...`) and passes linter checks (`go vet ./...`).
3. **Commit clearly**:
   Use conventional commits where possible (e.g., `feat:`, `fix:`, `docs:`, `refactor:`).
4. **Open a Pull Request**:
   Explain the rationale, especially referencing FACS literature or architectural considerations where applicable.

---

## 📜 Code of Conduct & Copyright

* **No Copyrighted Material**: Do not commit or submit pull requests containing copyrighted proprietary FACS test images, scans, or restricted manual pages. Only contribute code, documentation, and tooling for user-provided datasets.
* Respectful collaboration is expected at all times across developers, academics, and contributors.