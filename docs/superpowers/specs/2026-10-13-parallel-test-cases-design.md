# Parallel test cases (`--parallel-level=test`) — Design

Date: 2026-10-13
Status: Approved for planning
Builds on: #906 (opt-in parallel execution of test suites)

## Goal

Let users run individual test cases (`it:` jobs) concurrently, not only whole
suites. This speeds up charts whose tests are concentrated in a few large suite
files, in suites that share one `.snap` file, or in a skewed layout where one
huge suite dominates wall-clock time — cases where suite-level parallelism
(#906) gives little or no speedup.

## Success criteria

- `--parallel --parallel-level=test` produces output, counts, formatter reports
  and `.snap` files identical to a sequential run on the same input.
- Total concurrent test-case executions never exceed `--max-workers`.
- Existing behaviour of `--parallel` (no `--parallel-level`) is unchanged.
- The full test suite passes under `-race`.

## Non-goals

- Changing the default of `--parallel-level` (may be revisited in a later release).
- Running suites that share a snapshot file concurrently with each other.
- Parallelism inside a single test case (assertions stay sequential).

## CLI

| Flag | Type | Default | Meaning |
|---|---|---|---|
| `--parallel` | bool | `false` | Unchanged. Enables parallel execution. |
| `--parallel-level` | string | `suite` | `suite` = #906 behaviour; `test` = test cases from all suites share one worker pool. Only used with `--parallel`. |
| `--max-workers` | int | `0` | Unchanged. Pool size; `0` = number of CPU cores. |

Rules:

- Any value other than `suite` or `test` is a CLI error at startup.
- `--parallel-level=test` without `--parallel` logs a warning and runs sequentially.
- `--debugPlugin` forces sequential execution regardless of both flags (existing
  warning is kept).

## Execution model (`--parallel-level=test`)

### Single shared pool

- One semaphore (`chan struct{}`) of size `effectiveWorkers(totalCases)` is created
  per chart run.
- Only **leaf work** — executing one test case — acquires a slot. Coordinators
  (suite-group goroutines, suite runners) never hold a slot. No code path holds a
  slot while waiting for another slot, so the pool cannot deadlock.
- The chart deep copy (`FullCopyV2Chart`) happens inside the slot, so at most
  `max-workers` chart clones exist at once.

### Runner (`pkg/unittest/test_runner.go`)

- `runV4SuitesOfChart` dispatches to: sequential, suite-level (#906, unchanged),
  or test-level (new).
- Test-level path:
  1. Group suites by snapshot file (existing `groupSuitesBySnapshotFile`).
  2. Start one goroutine per group, without a semaphore slot.
  3. Within a group, suites run sequentially (they share one `snapshot.Cache`, and
     commit order must be deterministic); each suite's cases fan out into the pool.
  4. The main goroutine waits on groups in discovery order and prints/counts/appends
     results exactly as the existing parallel path does (`done[i]` channels).
  5. Group-level `--failfast` across groups: once a suite fails, groups that have not
     started are skipped (as in #906). Because all groups start immediately at this
     level, the main goroutine additionally stops printing after the first suite
     result that triggers failfast — the same condition the sequential loop breaks
     on — so printed output and counts equal a sequential failfast run. Groups that
     were already in flight finish and store their own (complete) snapshot files;
     this matches #906 and is a documented limitation. The `suite` level keeps its
     current printing behaviour unchanged.

### Suite (`pkg/unittest/test_suite.go`, `runV4TestJobs`)

- New optional `pool` parameter. `pool == nil` keeps the current sequential loop;
  each case's snapshot record is committed immediately after the case runs.
- `pool != nil`:
  1. Start one goroutine per case. Right after acquiring a slot (so the freshest
     value is seen after waiting) it checks an atomic `stopAfter` index; if a
     lower-index case has already failed under `--failfast`, it releases the slot
     without running.
  2. Inside the slot: clone chart, `WithConfig`, `RunV4`, store the job result and
     the case's snapshot `Recorder` into preallocated `slots[idx]`, release slot.
  3. On a failure with `--failfast`, atomically lower `stopAfter` to `idx` (min).
  4. After `wg.Wait()`, walk `slots` in index order applying exactly the current
     loop's rules (`idx == 0` initialisation of `Pass`, skip handling, `FailFast`).
     Stop at the first failing index when `--failfast` is set.
  5. Commit snapshot recorders to the cache **in index order, only for kept cases**.
  6. Results for dropped cases are `nil`, matching how sequential failfast leaves
     later `jobResults` entries `nil` (`handleSuiteResult` already skips them).

## Snapshot model

Problem: concurrent cases mutate `Cache.current` and the counters; with failfast,
cases that are dropped from the report would still leave snapshot entries and
counts behind.

Solution: separate the read-only comparison from the recording.

- `snapshot.Recorder` (new, per test case):
  - `Compare(test, idx, content, opts...) *CompareResult` — reads `cache.cached`
    (immutable after `RestoreFromFile`) and `cache.IsUpdating`, computes the result
    returned to the assertion exactly as today, and buffers the value to save plus
    inserted/updated/current increments locally.
- `Cache.Commit(rec *Recorder)` — applies the buffered values to `current` and the
  counters. Called only from the suite's coordinator goroutine, so `Cache` needs no
  mutex.
- `Cache.Compare` keeps its signature and becomes "record and commit immediately",
  so existing callers and tests are unaffected.
- `orderedSnapshotComparer` (`test_job.go`) holds a `*Recorder` instead of a
  `*Cache`.

Consequences:

- Commit in index order makes duplicate `it:` names deterministic: the last one wins,
  as in sequential mode. No duplicate detection is needed.
- `-u` behaves identically; the `.snap` file is still written once per group after
  all its suites finish.
- `CountSnapshot` runs after `runV4TestJobs` returns, i.e. after commits, so suite
  snapshot counts only include kept cases.
- A group cut short by failfast is not stored (existing rule in `runSuiteGroup`).
  This rule is extended, **in all modes** (sequential, `suite`, `test`), so that a
  suite whose cases were cut short by failfast also marks the group as not
  completed. Today, a failfast inside the *last* suite of a group still stores the
  file, so snapshots of cases that never ran are treated as vanished and deleted.
  This is a bug fix (CHANGELOG entry) and keeps all three modes in parity.

## Error handling

- Invalid `--parallel-level`: command returns an error before running anything.
- Panics inside a test case are not newly recovered; behaviour matches the
  sequential path.
- Snapshot cache creation and store errors are reported as today (synthesized
  results that are printed and counted but not appended to formatter results).

## Testing

- **Parity:** for fixtures under `test/data/v3`, sequential, `suite` and `test`
  levels produce identical printed output, counts and formatter results.
- **Snapshots:** `-u` under `test` yields byte-identical `.snap` files; duplicate
  `it:` names resolve as in sequential mode; snapshot counts match.
- **Failfast:** `test`-level failfast output equals sequential failfast output;
  dropped cases leave no entries in the `.snap` file and no counts.
- **Concurrency:** a hook-based test proves cases actually overlap and that the
  number of concurrently running cases never exceeds `max-workers`.
- **Recorder unit tests:** `Recorder.Compare` + `Cache.Commit` equal the old
  `Cache.Compare` for insert, match, mismatch, update mode and regex options.
- **CLI:** invalid value errors; `test` without `--parallel` warns.
- Whole suite under `-race` (Makefile `unittest` target already uses it).

## Documentation

- README: describe `--parallel-level`, its interaction with `--parallel`,
  `--max-workers`, `--failfast` and `--debugPlugin`.
- CHANGELOG entries under the unreleased `1.3.0` section. No `plugin.yaml` version
  bump: the repository bumps versions in its separate "Prepare for release" commits.
