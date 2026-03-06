# Go + React Fullstack Application

A fullstack application demonstrating industrial-standard Go backend architecture paired with a modern React frontend.

## Project Structure

```
Go-language/
├── backend/                    # Go backend (Clean Architecture)
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         # Entry point, graceful shutdown
│   ├── internal/
│   │   ├── api/
│   │   │   ├── router.go       # Chi router + middleware
│   │   │   └── handlers/       # HTTP handlers
│   │   ├── config/             # Environment config
│   │   ├── domain/             # Core domain types
│   │   ├── repository/         # Data access layer
│   │   └── service/            # Business logic layer
│   └── go.mod
└── frontend/                   # React + Vite frontend
    ├── src/
    │   ├── App.jsx             # Main component
    │   ├── index.css           # Global styles (glassmorphism)
    │   └── lib/api.js          # Axios API client
    └── package.json
```

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
