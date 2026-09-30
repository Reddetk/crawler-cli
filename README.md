# crawler-cli

A small Go CLI that crawls HTML pages concurrently, extracts `<title>` and links, stays within each start URL's host, prevents exact-URL cycles, and writes the resulting page trees to JSON. A failed resource is logged and skipped without stopping sibling work.

## Build

The module currently declares Go 1.26. Create the ignored `resources` directory before starting because the default Zap log path is opened during application startup.

```powershell
New-Item -ItemType Directory -Force resources | Out-Null
go mod download
go build -o crawler-cli.exe ./cmd
```

## Run

`--urls` is required; durations use Go syntax. `--log` adds another log destination while `resources/crawler.log` remains the default destination.

```powershell
.\crawler-cli.exe ` 
  --urls https://example.com,https://www.bsuir.by/ `
  --depth 3 `
  --timeout 2m `
  --request-timeout 10s `
  --output resources/result.json `
  --log stdout
```

## Configuration

| Option | Default | Meaning |
| --- | ---: | --- |
| `--urls` | required | Comma-separated start URLs |
| `--depth` | `3` | Maximum link depth; `0` crawls only roots |
| `--timeout` | `2m` | Whole-run timeout |
| `--request-timeout` | `1m` | Timeout for one HTTP request |
| `--output` | `resources/result.json` | Result file |
| `--log` | — | Additional Zap output path |
| `MAXWORKERS` | `5` | Concurrent observations; use `1`–`10` |

## Runtime behavior

Each root owns a crawl tree and visited-URL set; child calls use goroutines while a shared channel semaphore bounds active observations. Redirects, non-HTML responses, non-200 responses, unreachable pages, and request timeouts are skipped. `Ctrl+C` stops queued and active work, prunes unfinished nodes, saves completed pages, and exits with code `130`; the overall timeout also saves a partial result and exits with code `1`.

## Output

The output is a JSON array of recursive page nodes; leaves always contain an empty `links` array.

```json
[
  {
    "resource": "https://example.com",
    "title": "Example Domain",
    "links": []
  }
]
```

## Architecture

`cmd` is the composition root; the primary CLI adapter owns flags, run lifecycle, reporting, and JSON persistence; `core` owns traversal, depth, deduplication, pruning, and concurrency; the secondary HTTP adapter performs requests, charset conversion, and HTML parsing; primary and secondary ports separate orchestration from I/O.

## Tests

The test suite uses port doubles and local HTTP servers to cover depth, host filtering, resource failures, cancellation, partial results, worker limits, parser behavior, legacy encodings, and ten cyclic graph shapes.

```powershell
go test ./...
go test -race -count=20 -shuffle=on ./internal/core
go test -v -run 'TestStartCrawl_CyclicGraphs' ./internal/core
go test -cover ./...
```

## Profiling

```powershell
go test ./internal/core -count=1 -run 'TestStartCrawl_CyclicGraphs' -trace=trace.out
go tool trace trace
```
