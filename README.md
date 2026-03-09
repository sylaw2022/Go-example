# Go + Next.js Fullstack Application

A fullstack application demonstrating industrial-standard Go backend architecture paired with a modern Next.js frontend.

## Project Structure

```
Go-language/
├── backend/                    # Go backend (Clean Architecture / DDD)
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         # Entry point, graceful shutdown
│   ├── internal/
│   │   ├── api/
│   │   │   ├── router.go       # Chi router + middleware
│   │   │   └── handlers/       # HTTP handlers
│   │   ├── config/             # Environment config
│   │   ├── domain/             # Core domain types & errors
│   │   ├── repository/         # Data access layer (Repository pattern)
│   │   └── service/            # Business logic layer
│   └── go.mod
├── frontend/                   # Next.js 14 App Router frontend
│   ├── app/
│   │   ├── globals.css         # Global styles (glassmorphism)
│   │   ├── layout.tsx          # Root layout
│   │   └── page.tsx            # Main page (Server Component)
│   ├── lib/
│   │   └── api.ts              # Typed fetch helper (server-side)
│   ├── types/
│   │   └── user.ts             # TypeScript User interface
│   └── next.config.ts          # API proxy config
└── frontend-old/               # Preserved original Vite + React frontend
```

## Running the App

```bash
# Terminal 1 — Go backend
cd backend && PORT=9090 go run ./cmd/server/main.go

# Terminal 2 — Next.js frontend (http://localhost:3000)
cd frontend && npm run dev
```

---

## Building the Backend

### Development (run without compiling)
```bash
cd backend
PORT=9090 go run ./cmd/server
```

### Compile to a Binary
```bash
cd backend
go build -o ./bin/server ./cmd/server
PORT=9090 ./bin/server
```

### Production Build (smaller binary, no debug symbols)
```bash
go build -ldflags="-s -w" -o ./bin/server ./cmd/server
```

| Flag | Effect |
|---|---|
| `-s` | Strips symbol table |
| `-w` | Strips DWARF debug info |
| Result | ~30-40% smaller binary |

### Cross-Compile for Other Platforms
```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o ./bin/server.exe ./cmd/server

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o ./bin/server-mac ./cmd/server
```

Go compiles to a **self-contained native binary** with zero external dependencies — no runtime, no `libc`, no Docker required to run it.

### Makefile
Add a `Makefile` to `backend/` for convenience:

```makefile
.PHONY: build run dev clean

build:
	go build -ldflags="-s -w" -o ./bin/server ./cmd/server

run: build
	PORT=9090 ./bin/server

dev:
	PORT=9090 go run ./cmd/server

clean:
	rm -rf ./bin/
```

Then use `make dev`, `make build`, or `make run`.

---

## Q&A: Concepts Covered in This Session


### 1. What is `github.com/sylaw/fullstack-app/internal/api` in the import?

These are **Go import paths**. Go uses a URL-like naming convention for module and package identification.

- The first part (`github.com/sylaw/fullstack-app`) is the **module name** declared in `go.mod`. When Go sees this prefix in an import, it knows to look at local files — not the internet.
- The second part (`/internal/api`) is the **relative directory path** inside the project root.
- Using a URL (`github.com/username/repo`) is industrial standard so the paths are globally unique if the code is ever published.

**Go's resolution rule:**
- If the import URL **matches** the `go.mod` module name → look locally on disk.
- If the import URL **doesn't match** the `go.mod` module name → download from the internet.

---

### 2. What does "look in the root directory of current project" mean?

The `go.mod` file defines the **project root**. When you run `go mod init github.com/sylaw/fullstack-app` inside the `backend/` folder, that file tells Go:
> "The folder I live in is the top-level of the module called `github.com/sylaw/fullstack-app`."

So `import "github.com/sylaw/fullstack-app/internal/service"` resolves to the physical path `/home/sylaw/Go-language/backend/internal/service/`.

---

### 3. Why is `internal/` special?

Any code inside a folder named `internal/` is **private to that module**. Go's compiler prevents any external module from importing it. This protects your core business logic from being accidentally used as a public library.

---

### 4. How does the overall project data flow work?

```
React (Browser)
  → Axios HTTP GET /api/v1/users
  → Go Chi Router (internal/api/router.go)
  → UserHandler.GetAll() (internal/api/handlers/)
  → UserService.GetUsers() (internal/service/)
  → UserRepository.GetAll() (internal/repository/)
  → [Data Source / In-Memory Store]
  → Returns []domain.User up the chain
  → Handler serializes to JSON
  → HTTP 200 Response
  → React renders user cards
```

---

### 5. When is the Go Runtime used?

The Go Runtime is always active while the server runs. Key uses:

- **Per HTTP Request:** Go's `net/http` server automatically spawns a new **goroutine** for every incoming HTTP connection. Handlers run inside these goroutines — you write code synchronously and Go handles concurrency automatically.
- **Explicit `go` keyword:** Used in `main.go` for the graceful shutdown goroutine, so the main thread can proceed to `ListenAndServe()` without blocking.

---

### 6. How does the Node.js Event Loop work, and what is its issue?

Node.js is **single-threaded** — only one OS thread executes JavaScript. It handles I/O concurrency via the **Event Loop**:

1. An I/O call (e.g. database query) is handed off to the OS (`epoll`/`kqueue`) or libuv's Thread Pool.
2. The main thread immediately moves on to serve other requests.
3. When I/O completes, the result is placed in the **Event Queue**.
4. The Event Loop picks it up and runs the callback.

**The Issue — CPU-Bound Tasks:**
If JavaScript executes heavy computation (image processing, password hashing loop), the single main thread is fully occupied. All other pending requests are completely blocked until the computation finishes. This is the fundamental scaling bottleneck of Node.js for CPU-intensive workloads.

---

### 7. Does Node.js spawn threads for I/O?

**Usually No — it uses OS non-blocking I/O:**
For network operations, Node.js uses `epoll` (Linux) / `kqueue` (macOS). The OS monitors sockets using hardware interrupts — no thread is sleeping and waiting. When data arrives, the OS notifies Node.js's Event Queue.

**Sometimes Yes — the libuv Thread Pool:**
For operations without a non-blocking OS equivalent (file system reads, DNS lookups, crypto operations), Node uses a **hidden thread pool of 4 C++ threads**. If all 4 threads are busy, the 5th request queues and waits.

---

### 8. Is Node.js suitable for large-scale backends?

**Yes — for I/O-bound workloads** (standard CRUD APIs, real-time chat, WebSockets). Netflix, Uber, and LinkedIn use Node.js at massive scale.

**No — for CPU-bound workloads** (video transcoding, complex math, real-time message routing at high throughput). Discord rewrote their message routing from Node.js to **Go** for this reason.

---

### 9. Since Go uses goroutine-per-request, will CPU-intensive work in one goroutine block others?

**No**, because of the **Go Preemptive Scheduler**. Every ~10 milliseconds, the Go Scheduler checks each goroutine. If one has hogged a CPU core, the scheduler **forcibly pauses** it, puts it back in the run queue, and lets another goroutine run. All goroutines eventually get fair CPU time slices.

---

### 10. Does Go implement its own virtual thread and scheduler on top of the OS?

Yes. Every compiled Go binary embeds the **Go Runtime**, containing:

1. **M:N Scheduler:** Maps N goroutines (virtual threads) onto M OS threads (one per CPU core), using work-stealing for load balancing.
2. **Network Poller:** Uses `epoll`/`kqueue` to handle network I/O without blocking goroutines.

---

### 11. Is the Go Runtime a single-threaded OS process like Node.js?

No. The Go Runtime creates **as many OS threads as you have CPU cores** by default (e.g., 8 OS threads on an 8-core machine). This means Go can use **100% of all CPU cores simultaneously**, whereas Node.js can only ever max out 1 core.

---

### 12. Can a single OS thread run on multiple CPU cores simultaneously?

No. A single OS thread is one sequential stream of instructions. At any nanosecond, one CPU core executes instructions from exactly one thread. The **Linux Scheduler** time-slices hundreds of threads across your cores in 1-4ms intervals, creating the illusion of parallelism.

---

### 13. How does the Go Scheduler assign goroutines to OS threads? (M:P:G Model)

- **G (Goroutine):** Lightweight virtual thread you create with `go`.
- **M (Machine):** A real OS Thread.
- **P (Processor):** A logical context with a Local Run Queue of ~256 goroutines.

Each `M` is attached to a `P`. The `M` continuously pops goroutines from its `P`'s local queue and runs them. If a `P`'s queue is empty, it **steals half the goroutines** from another busy `P`'s queue (Work Stealing), ensuring no CPU core sits idle.

---

### 14. How does an OS Thread know how to run an assigned goroutine?

Each Goroutine (`G`) in memory stores:
- **Program Counter (PC):** Memory address of the goroutine's next instruction.
- **Stack Pointer (SP):** Pointer to the goroutine's private 2KB stack in RAM.
- **`gobuf`:** Saved CPU register state for when the goroutine was paused.

To execute a goroutine, the OS Thread calls the `gogo` Assembly function which:
1. Loads the goroutine's SP into the physical CPU stack register.
2. Restores saved registers from `gobuf`.
3. Executes `JMP` to the goroutine's PC.

---

### 15. Does Go use Assembly for context switching?

Yes. The Go Runtime includes hand-written Assembly (`.s` files) for each CPU architecture (amd64, arm64, etc.).

- **`gogo`**: Loads a goroutine's state into the CPU and jumps into it.
- **`mcall`**: Saves the current CPU state back into the goroutine's `gobuf` and returns to the scheduler.

This Assembly runs in **User Space (Ring 3)** — no kernel call required.

---

### 16. How can Ring 3 code manipulate CPU registers without a kernel fault?

General-purpose registers (`RSP`, `RIP`, `RAX`, `RBX`, etc.) are freely accessible by any Ring 3 program. Every function call in every program uses them.

**Only privileged registers are protected:**
| Register | Purpose | Ring Required |
|---|---|---|
| `CR3` | Page table pointer | Ring 0 only |
| `CR0` | CPU mode control | Ring 0 only |
| `GDTR/IDTR` | Descriptor tables | Ring 0 only |

Go only manipulates general-purpose registers during goroutine switching, so no privilege fault is triggered. Goroutines also share the same virtual address space (same `CR3`), so no page table switch is needed — unlike the Linux Kernel switching OS threads.

---

### 17. What happens if Go Runtime writes an invalid address into the PC?

1. The CPU attempts to fetch instructions from the invalid address.
2. The MMU checks the page table — the address is likely unmapped.
3. The CPU fires a **Page Fault (#PF)** or **General Protection Fault (#GP)**.
4. Linux catches the fault via an Interrupt Service Routine (ISR).
5. Linux sends **SIGSEGV** to the Go process.
6. The **entire Go server process crashes** — not just the affected goroutine.

---

### 18. How does the Go Scheduler compare to the Linux Scheduler?

Both use the **save-and-restore context switch** (written in Assembly). The difference is cost:

| | Linux Thread Switch | Go Goroutine Switch |
|---|---|---|
| **Runs in** | Kernel Space (Ring 0) | User Space (Ring 3) |
| **Registers saved** | ~200 (including SIMD/AVX) | ~6 general-purpose only |
| **Page table flush** | Yes (CR3 changes) | No (same address space) |
| **Ring transition** | User → Kernel → User | Stays in Ring 3 |
| **Cost** | ~1-2 microseconds | ~0.2 microseconds |

---

### 19. What is the difference between Go and Rust?

| Feature | Go | Rust |
|---|---|---|
| **Memory** | Garbage Collected | Ownership system (compile-time) |
| **Runtime** | Yes (GC + Scheduler) | None |
| **GC Pauses** | Sub-millisecond pauses | Zero — impossible |
| **Concurrency** | Goroutines (built-in) | async/await via Tokio |
| **Learning curve** | Gentle | Very steep |
| **Best for** | Web APIs, microservices, DevOps tooling | OS components, real-time systems, embedded |

---

### 20. Isn't the Linux Kernel written in C, not Rust?

C is the primary, dominant language of the Linux Kernel. However, since **Linux Kernel 6.1 (December 2022)**, Rust was officially accepted as a second supported language for writing **new** kernel drivers. The motivation is memory safety — memory bugs (buffer overflows, use-after-free) cause a large percentage of kernel CVEs. Rust's compile-time ownership rules eliminate this class of bugs entirely.

---

### 21. What is the difference between Android and the Linux Kernel?

Android is a full OS **built on top of the Linux Kernel**. The Linux Kernel is only the lowest layer.

```
Your Android App (Java/Kotlin)
Android Framework (Activity Manager, etc.)
ART (Android Runtime)
HAL (Hardware Abstraction Layer)
Modified Linux Kernel   ← Shared foundation
```

Key Android additions to the standard Linux Kernel:
- **Binder IPC** — fast cross-process RPC driver.
- **Wakelocks** — mobile battery management.
- **Low Memory Killer** — aggressively kills background apps on low RAM.
- **Zygote** — app fork model for fast startup.

---

### 22. Does Android use the same file system drivers as Linux when calling `fopen()`?

Yes. `fopen()` on Android follows the same kernel path as standard Linux:

```
App → Java API → JNI → Bionic libc → open() syscall → VFS Layer → FS Driver (ext4/F2FS) → Block Device
```

The difference: Android uses **Bionic libc** (lightweight, BSD-licensed) instead of **glibc**. Both call the same Linux kernel `open()` syscall, so from the kernel's perspective they are identical.

---

### 23. What is Android Binder, and is it a hardware device driver?

Binder is not a hardware device driver. It is a **cross-process IPC (Inter-Process Communication) mechanism**.

Android apps run in sandboxed processes. They cannot directly access hardware. When an app needs the camera, it must ask the `CameraService` (running in the privileged `system_server` process) to access the hardware on its behalf. **Binder is the channel** that allows this cross-process method call.

```
Your App Process → Binder IPC → system_server (CameraService) → /dev/video0 (standard Linux driver)
```

---

### 24. What is the difference between Binder and Linux Message Queue?

| Feature | Linux Message Queue | Android Binder |
|---|---|---|
| **Style** | One-way async queue | Two-way synchronous RPC |
| **Data copies** | 2 (user → kernel → user) | 1 (via `mmap`) |
| **Caller identity** | Not provided | Built-in UID/PID verification |
| **Object passing** | Raw bytes only | Full object/interface references |
| **Security** | File system permissions | Fine-grained Android permission system |

Binder was created because Message Queues are too slow and complex for the hundreds of cross-process calls Android makes per second. Binder makes a cross-process call feel like a normal local method call using **AIDL (Android Interface Definition Language)**, which auto-generates all serialization code.

---

### 25. What does the graceful shutdown goroutine do in `main.go`?

The graceful shutdown goroutine ensures that when the server receives a stop signal (e.g., `Ctrl+C` or `SIGTERM`), it finishes processing active requests before exiting rather than rudely cutting off users mid-request.

```go
go func() {
    <-sig                  // blocks until OS signal arrives
    server.Shutdown(ctx)   // stops new connections, lets current ones finish
    serverStopCtx()        // signals main() it's safe to exit
}()
```

1. **`<-sig`** blocks the goroutine until an OS interrupt signal arrives.
2. A 30-second context (`context.WithTimeout`) is created — the server must finish all active requests within this window.
3. A watchdog goroutine monitors the deadline. If 30 seconds elapse before completion, `log.Fatal` forces the process to exit.
4. `server.Shutdown(ctx)` stops accepting new connections but lets in-progress requests complete.
5. `serverStopCtx()` notifies `main()` that cleanup is done and the process may exit cleanly.

---

### 26. What is `NewInMemoryUserRepository()`?

It is a **constructor function** that creates a fake, in-memory database pre-populated with test data. Instead of connecting to PostgreSQL, it stores users in a Go `map` in RAM.

```go
func NewInMemoryUserRepository() *InMemoryUserRepository {
    return &InMemoryUserRepository{
        users: map[int]domain.User{
            1: {ID: 1, Name: "Alice Smith",   Email: "alice@example.com"},
            2: {ID: 2, Name: "Bob Jones",     Email: "bob@example.com"},
            3: {ID: 3, Name: "Charlie Brown", Email: "charlie@example.com"},
        },
    }
}
```

- **Fast prototyping:** No database setup required.
- **Unit testing:** Tests run instantly with no external dependencies.
- **Easy swap:** Replace with `NewPostgresUserRepository()` in `main.go` and the rest of the app is unchanged.

The struct also embeds a `sync.RWMutex` to safely handle concurrent HTTP requests reading/writing the map simultaneously.

---

### 27. What is `users map[int]domain.User`?

This is a struct field declaration defining a **hashmap (dictionary)** where:
- **Key:** `int` — the user's ID
- **Value:** `domain.User` — the full user struct

It enables O(1) lookup by ID (`r.users[42]`) rather than scanning a slice sequentially. When `GetAll()` is called, the map is converted to a `[]domain.User` slice for the HTTP response, because JSON APIs return ordered arrays.

---

### 28. What is the `UserRepository` interface?

```go
type UserRepository interface {
    GetAll() ([]domain.User, error)
    GetByID(id int) (domain.User, error)
}
```

An interface is a **contract**. Any Go type that implements these two methods automatically satisfies `UserRepository` — no `implements` keyword needed.

This enables the **Repository Pattern**: the service layer depends only on the interface, not on any specific database implementation. Swapping `InMemoryUserRepository` for `PostgresUserRepository` requires changing only one line in `main.go`.

---

### 29. What is a Go method receiver like `func (r *InMemoryUserRepository) GetAll()`?

In Go, there are no classes. Instead, you attach methods to types using a **receiver** — Go's explicit equivalent of `this`/`self`:

```go
func (r *InMemoryUserRepository) GetAll() ([]domain.User, error) {
    // r is the struct instance
}
```

- `r` is a pointer receiver (`*`) — it accesses the real struct in memory, not a copy.
- This is what satisfies the `UserRepository` interface. Standalone functions cannot satisfy interfaces; only methods attached via receivers can.

---

### 30. What is the difference between `[]domain.User` and `map[int]domain.User`?

| | `[]domain.User` (Slice) | `map[int]domain.User` (Map) |
|---|---|---|
| **Access** | By position: `users[0]` | By key: `users[42]` |
| **Order** | Always preserved | Not guaranteed |
| **Lookup by ID** | O(n) — must scan all | O(1) — direct jump |
| **Use when** | Returning a list to a client | Storing data for fast ID lookup |

The repository stores data as a **map** for fast lookups but returns a **slice** for the HTTP response because JSON clients expect an ordered array.

---

### 31. What is `var ErrUserNotFound = errors.New("user not found")`?

This is a **sentinel error** — a named, package-level error value that callers can precisely identify using `errors.Is()`.

```go
// In the repository
return domain.User{}, domain.ErrUserNotFound

// In the handler — identify the specific error
if errors.Is(err, domain.ErrUserNotFound) {
    http.Error(w, "not found", http.StatusNotFound) // 404
    return
}
```

Sentinel errors are declared in the `domain` package because they belong to the business domain (not to the HTTP or database layers), preventing circular imports.

---

### 32. How does Go know when `exists` is `false` in `user, exists := r.users[id]`?

This is Go's built-in **"comma ok" idiom** for map lookups. When you access a map with two variables, the second bool is automatically set by the runtime:

- `true`: the key was found, `user` holds the value.
- `false`: the key does not exist, `user` is the zero value (`domain.User{}`).

You cannot rely on checking if the value is empty because zero-value structs are valid data (a user with ID 0 and empty name is possible). The `exists` bool is the only reliable indicator.

---

### 33. What do `r.mu.RLock()` and `defer r.mu.RUnlock()` do?

A `sync.RWMutex` protects the map from concurrent read/write corruption. Go maps are not thread-safe — two simultaneous goroutines accessing the same map can crash the program with `fatal error: concurrent map read and map write`.

- **`RLock()`:** Acquires a read lock. Multiple goroutines can hold a read lock simultaneously, but it blocks any writer.
- **`defer RUnlock()`:** Releases the lock when the function returns, no matter what — including on error return paths.

For write operations (`Lock()`/`Unlock()`), all concurrent readers and writers are blocked.

---

### 34. What does `NewUserService(repo repository.UserRepository)` demonstrate?

This constructor demonstrates **Dependency Injection** — the repository is passed in from the outside rather than created inside the service:

```go
func NewUserService(repo repository.UserRepository) UserService {
    return &userServiceImpl{ repo: repo }
}
```

Key design points:
1. **Accepts an interface**, not a concrete type — any `UserRepository` implementation works.
2. **Returns an interface** (`UserService`) — callers cannot access the concrete `userServiceImpl` struct.
3. **`userServiceImpl` is lowercase (unexported)** — only accessible through the interface.

This is wired in `main.go`:
```go
repo    := repository.NewInMemoryUserRepository()
service := service.NewUserService(repo)
handler := handler.NewUserHandler(service)
```

---

### 35. What does `SetupRouter` do?

`SetupRouter` creates and configures the `chi` HTTP router — the traffic controller for all incoming web requests.

**Middleware chain (runs on every request):**
1. `RequestID` — stamps each request with a unique UUID for distributed tracing.
2. `RealIP` — extracts the real client IP from `X-Forwarded-For` headers (behind load balancers).
3. `Logger` — logs method, URL, response status, and duration automatically.
4. `Recoverer` — catches Go panics and returns HTTP 500 instead of crashing the server.
5. **CORS middleware** — sets `Access-Control-Allow-Origin` headers and handles preflight `OPTIONS` requests.

**Routes registered:**
| Method | Path | Handler |
|---|---|---|
| GET | `/health` | `HealthCheck` |
| GET | `/api/v1/users` | `GetAll` |
| GET | `/api/v1/users/{id}` | `GetByID` |

---

### 36. How does `RequestID` middleware help debug production issues?

Every request is stamped with a UUID. The `Logger` records it alongside the timestamp, URL, and status code:

```
2026-03-09 15:00:04  GET  /api/v1/users/99  404  reqID=a3f9c1b2
```

When a user reports "my request failed at 3pm," they provide their reference ID (shown in the error dialog). You search logs for `a3f9c1b2` and instantly see the full request lifecycle — every log line your handlers emitted is tagged with the same ID.

---

### 37. How does `RealIP` middleware work?

Production servers sit behind load balancers/proxies (Nginx, Cloudflare). Without `RealIP`, every request appears to originate from the proxy's internal IP (`127.0.0.1`).

`RealIP` reads the `X-Real-IP` or `X-Forwarded-For` header set by the proxy and overwrites `r.RemoteAddr` with the actual client IP. This is required for rate limiting, geo-blocking, audit logs, and fraud detection to work correctly.

---

### 38. How does the CORS middleware work?

Browsers enforce the **Same-Origin Policy**: JavaScript on `http://localhost:5173` is blocked from reading responses from `http://localhost:8080` unless the backend explicitly allows it. Origin = protocol + domain + port; different port = different origin.

The CORS middleware adds three headers to every response:
- `Access-Control-Allow-Origin: *` — which frontend domains are trusted.
- `Access-Control-Allow-Methods` — which HTTP verbs are permitted.
- `Access-Control-Allow-Headers` — which request headers (e.g., `Authorization`) are permitted.

Browsers send a **preflight `OPTIONS` request** before any cross-origin POST/PUT with custom headers. The middleware responds `200 OK` immediately without invoking business logic, then the browser sends the real request.

> In production, replace `"*"` with your exact front-end URL (e.g., `"https://yourapp.com"`) to prevent other websites from exploiting your users' authenticated sessions (CSRF).

---

### 39. Why doesn't `HealthCheck` return the HTTP response?

In Go's `net/http`, `w http.ResponseWriter` is not a value you return — it is a **live pipe directly connected to the client's TCP socket**. Writing to `w` sends data to the browser in real time:

```go
w.Header().Set("Content-Type", "application/json")  // writes headers to buffer
w.WriteHeader(http.StatusOK)                         // sends "HTTP/1.1 200 OK"
json.NewEncoder(w).Encode(...)                       // streams JSON body
```

When the handler function returns, the Go HTTP server flushes any remaining data and closes the response. This design enables streaming large responses chunk-by-chunk without buffering the entire payload in memory.

---

### 40. What is `chi`?

`chi` is a lightweight, idiomatic Go HTTP router. Go's built-in `net/http` has no URL parameter extraction or route grouping. `chi` adds:
- **URL parameters:** `r.Get("/users/{id}", handler)` and `chi.URLParam(r, "id")`.
- **Route groups:** `r.Route("/api/v1", func(r chi.Router) { ... })`.
- **Method-specific routing:** `r.Get(...)`, `r.Post(...)`, `r.Delete(...)`.
- **Composable middleware:** `r.Use(...)`.

Unlike full frameworks (Gin, Echo), `chi` uses standard `http.HandlerFunc` throughout — your code stays idiomatic Go and your handlers are 100% compatible with the standard library.

---

### 41. What is `context.WithCancel(context.Background())`?

```go
serverCtx, serverStopCtx := context.WithCancel(context.Background())
```

- `context.Background()` is the root context — it never expires and never cancels.
- `context.WithCancel(...)` wraps it and returns: a child **context** (`serverCtx`) and a **cancel function** (`serverStopCtx`).
- Calling `serverStopCtx()` closes `serverCtx.Done()`, signalling all code watching this context to stop.
- Contexts form a tree: cancelling a parent automatically cancels all children (including the 30-second shutdown timeout context derived from `serverCtx`).

---

### 42. What is `ctx.Done()` and how does `<-ctx.Done()` work?

`ctx.Done()` returns a **read-only channel** (`<-chan struct{}`) that is **closed** when the context is cancelled or its deadline expires. Receiving from a closed channel in Go unblocks immediately and returns the zero value — this is used as a broadcast signal to all goroutines waiting on `<-ctx.Done()`.

```go
<-serverCtx.Done()   // main() sleeps here until serverStopCtx() is called
```

Two cancellation reasons are distinguishable via `ctx.Err()`:
- `context.Canceled` — manually cancelled (clean shutdown).
- `context.DeadlineExceeded` — timeout elapsed (force-kill path).

---

### 43. How do Go's Native Syscalls work, bypassing `libc`?

Most languages (C, Python, Node.js) talk to the OS through `libc` (e.g., `glibc` on Linux). Go **bypasses `libc` entirely** on Linux by writing raw Assembly instructions that trap directly into the Linux Kernel.

**How a syscall works:**
1. Go loads the syscall ID and arguments into CPU registers:
   - `RAX` = syscall number (e.g., `1` = `sys_write`)
   - `RDI/RSI/RDX` = arguments
2. Executes the single `SYSCALL` assembly instruction.
3. The CPU switches from **Ring 3 (User Mode)** to **Ring 0 (Kernel Mode)**.
4. The kernel executes the request and returns control via `SYSRET`.

**Why this matters:**
- Go binaries have **zero dynamic library dependencies** on Linux.
- They run inside a completely empty Docker container (`FROM scratch`).
- Cross-compilation works without any toolchain: `GOOS=linux GOARCH=amd64 go build` on a Mac produces a perfect Linux binary.

> On macOS and Windows, Go routes through the OS's standard library because Apple and Microsoft change their internal syscall IDs without compatibility guarantees.

---

### 44. What is Domain-Driven Design (DDD)?

DDD is a software architecture philosophy where your **code structure mirrors the real-world business problem** (the "domain") rather than technical concerns.

| DDD Concept | Your Code |
|---|---|
| **Entity** | `domain.User` — has a unique identity (`ID`) that persists over time |
| **Value Object** | `Money{Amount, Currency}` — defined by its value, immutable |
| **Repository Interface** | `UserRepository` in `repository/` — domain defines the contract |
| **Repository Implementation** | `InMemoryUserRepository` / `PostgresUserRepository` — infrastructure detail |
| **Sentinel Error** | `domain.ErrUserNotFound` — domain-owned error |
| **Service** | `UserService` — business logic coordination |
| **Delivery Mechanism** | HTTP handlers — replaceable transport layer |

The folder structure already enforces DDD: the `domain` package has no external imports, and all other layers import from it — never the reverse.

---

### 45. Do you need to understand Go syntax to be an expert?

Knowing syntax is necessary but is only the **floor**, not the ceiling. Go's syntax is deliberately minimal (the full spec can be read in a few hours). What separates experts is depth in four areas:

1. **Goroutines & Channels** — when NOT to use them; designing channel communication to avoid deadlocks.
2. **Error handling** — wrapping errors with `fmt.Errorf("...: %w", err)`, `errors.Is()`, `errors.As()`, sentinel errors.
3. **Interfaces** — small, focused interfaces (`io.Reader`) vs bloated ones; the zero-value rule.
4. **`context.Context`** — propagating deadlines and cancellation through all layers correctly.

---

### 46. Do Go design principles apply to other languages?

Almost entirely. The best design principles are **language-agnostic**:

| Go Concept | Universal Principle | Other Languages |
|---|---|---|
| Small interfaces | Program to abstractions | Java `Interface`, TypeScript `interface`, Python `Protocol` |
| Repository pattern | Separate data access from logic | Every language |
| Layered architecture | Dependency flows inward | Clean Architecture, Hexagonal Architecture |
| Explicit errors | Always handle failure paths | Rust `Result<T,E>`, Swift `throws` |
| Composition over inheritance | Has-a over is-a | Go uses struct embedding; applies everywhere |

Go is one of the best languages to learn design principles because it **enforces** good habits by refusing to compile unused imports, having no inheritance, and returning errors explicitly.

---

### 47. Why does the Next.js migration eliminate CORS?

In the Vite+React app, the **browser** called the Go backend directly:
```
Browser (localhost:5173) → Go API (localhost:8080)  → CORS required
```

In Next.js with **React Server Components**, the HTTP request is made by the Next.js **server process**, not the browser:
```
Browser → Next.js server (localhost:3000) → Go API (localhost:9090)  → No CORS
```

CORS is a browser security check. Server-to-server communication has no browser involved and never triggers a CORS check. The browser only ever talks to the Next.js server — one origin, zero CORS issues.

