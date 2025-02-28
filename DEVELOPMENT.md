# Cinea Media Server - Development Guidelines

This document outlines the development guidelines for Cinea, a personal media server project. It prioritizes a pragmatic, maintainable approach suitable for a solo developer, balancing best practices with the realities of a personal project.

## 1. Core Principles (Solo Developer Focus)

*   **Maintainability (Top Priority):** The code must be easy for *you* to understand, even after periods of inactivity. Clear structure, comments, and this document are crucial.
*   **Reliability:** Cinea should be stable and handle errors gracefully. Unexpected crashes or data loss should be avoided.
*   **Scalability (Reasonable):** Design for a reasonable amount of media and users, without over-engineering for massive scale. Good concurrency practices are key, but avoid premature optimization.
*   **Pragmatic Testing:** Focus on *minimal viable testing* to catch the most likely bugs and prevent major regressions.
*   **Observability:** Thorough logging is essential for debugging and understanding application behavior.

## 2. Concurrency Management (Avoid Over-Engineering)

**Principle:** Use concurrency judiciously. Unbounded goroutines are a common source of problems. Prioritize simplicity and correctness.

**Implementation:**

*   **Worker Pools (For Key Tasks):** Use worker pools for:
    *   **Media Scanning:** Limit concurrent file system operations.
    *   **Transcoding (If Applicable):** Control simultaneous transcoding processes.
    *   **Metadata Fetching:** Limit concurrent requests to external providers.

    A custom worker pool implementation (see Appendix A) is likely sufficient. `golang.org/x/sync/errgroup` can also be helpful.

*   **Bounded Concurrency (Semaphores):** Use a semaphore (buffered channel) for fine-grained control over resource access (e.g., a specific database connection).  See example in original documentation.

*   **Focus on Common Scenarios:** Don't over-optimize for extreme edge cases.

## 3. Context Management (Essential for Graceful Shutdown)

**Principle:** `context.Context` is crucial for handling cancellations and timeouts, especially for I/O.

**Implementation:**

*   **Universal Context:** All functions that perform I/O, network requests, or significant processing *must* accept a `context.Context`.
*   **Timeouts:** Set reasonable timeouts for all network operations. Use `context.WithTimeout` or `context.WithDeadline`.
*   **Cancellation:** Implement graceful shutdown by propagating cancellation signals through the context. This is *critical* for a server application.
*   **Periodic Checks:** In long loops, check `ctx.Done()` to see if cancellation has been requested.
*   **Don't store in structs:** Pass `context.Context` explicitly.

## 4. Logging (Your Best Debugging Tool)

**Principle:** Log extensively, but strategically. Good logs are invaluable.

**Implementation:**

*   **Structured Logging:** Strongly consider `zap` or `logrus`. JSON-formatted logs are easier to analyze.
*   **Log Levels:** Use levels effectively:
    *   **Debug:** Detailed information (enable only when debugging).
    *   **Info:** General progress and status updates.
    *   **Warning:** Non-critical errors or potential problems.
    *   **Error:** Errors that prevent an operation from completing.
    *   **Fatal:** Critical errors that force shutdown.
*   **Contextual Information:** Always include:
    *   Usernames (if applicable)
    *   Media item IDs/paths
    *   Request IDs
    *   Error messages
*   **Log *Before* Operations:** Log your *intent* before attempting an operation that might fail.
*   **Privacy:** *Never* log PII or sensitive data.

## 5. Testing (Pragmatic and Minimal Viable)

**Principle:** Avoid "perfect testing." Focus on *minimal viable testing* to catch obvious bugs and prevent major regressions.

**Implementation (Prioritized):**

1.  **Error Handling Tests (Non-Negotiable):** For *every* function that returns an error, write a simple test to ensure the error is returned correctly. Use table-driven tests.  (See Appendix B for example).
2.  **Core Data Structure Validation (High Value):** Write basic tests for your `model` package to validate input and handle invalid input.
3.  **Critical Function Tests (Selective):** Choose *2-3 of the most critical functions* and write basic tests for the most common success/failure scenarios.
4.  **Race Detection (`go test -race`):** *Always* run your tests with the race detector.
5.  **NO TDD (Unless You Want To):** Write tests *after* coding, focusing on the priorities above.
6.  **"Good Enough" is Good Enough:** Don't aim for perfect coverage.

## 6. Error Handling (Be Explicit)

**Principle:** Handle errors meticulously. Don't ignore them, and provide clear, informative error messages.

**Implementation:**

*   **Check *Every* Error:** No exceptions.
*   **Wrap Errors:** Use `fmt.Errorf` with `%w` to wrap errors, adding context.
*   **Custom Error Types:** Define custom error types for specific situations (e.g., `ErrMediaNotFound`).
*   **Return Errors:** Prefer returning errors to panicking.
*   **Log Errors:** Log all errors, including the full error chain.

## 7. Project Structure (Keep It Organized)

*   **Modular Design:** Organize code into logical packages. A suggested structure:
    ```
    cinea/
        cmd/cinea/  (main package)
        internal/
            config/
            database/
            metadata/
            scanner/
            transcoder/ (if applicable)
            web/       (HTTP handlers, API routes)
            model/     (data structures)
            util/
        pkg/       (Reusable public packages, if any)
    ```
*   **Consistent Naming:** Follow consistent naming conventions.

## 8. Dependencies

*   **Manage Dependencies:** Use Go modules (`go mod`).
*   **Minimize Dependencies:** Be mindful of the number of external dependencies.

## 9. CI/CD

*   **Use CI/CD:** Automate testing and (potentially) deployment. GitHub Actions is a good option.

## Appendix

### A. Simple Channel-Based Worker Pool (Illustrative)
```Go

	type Task func() error

	func WorkerPool(numWorkers int, taskChan <-chan Task) {
		var wg sync.WaitGroup
		wg.Add(numWorkers)

		for i := 0; i < numWorkers; i++ {
			go func() {
				defer wg.Done()
				for task := range taskChan {
					if err := task(); err != nil {
						log.Printf("Worker error: %v", err) // Handle error
					}
				}
			}()
		}
	    wg.Wait()
	}

````

### B. Example Error Handling Test (Table-Driven)

Go

```
// Example:
func TestOpenFile(t *testing.T) {
    tests := []struct {
        path string
        err  error
    }{
        {"valid_file.txt", nil}, // Assuming valid_file.txt exists
        {"nonexistent_file.txt", os.ErrNotExist},
    }
    for _, tt := range tests {
        t.Run(tt.path, func(t *testing.T) {
            _, err := OpenFile(tt.path) // Your function
            if !errors.Is(err, tt.err) { // Use errors.Is for error comparison
                t.Errorf("OpenFile(%q) = %v; want %v", tt.path, err, tt.err)
            }
        })
    }
}
```
