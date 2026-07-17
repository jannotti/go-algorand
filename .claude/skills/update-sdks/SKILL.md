---
name: update-sdks
description: Propagate a go-algorand change out to the Algorand SDKs (Go, Java, JavaScript, Python) — regenerate the algod/indexer/kmd API specs, run the shared code generator, sync shared serialized types, and guide the manual Python edits. Use when the user asks to "update the SDKs", "regenerate SDK clients", or "propagate my go-algorand change to the SDKs".
argument-hint: [short description of what changed in go-algorand]
allowed-tools: [Bash, Read, Edit, Write, Grep, Glob]
---

# Update the Algorand SDKs after a go-algorand change

Take a change already made in `go-algorand` (a new/changed REST endpoint, a new
transaction/block field, a new consensus parameter, or a new shared serialized type)
and propagate it into the four sibling SDKs so their generated code and shared types
match.

This skill ships in the `go-algorand` repo and assumes it runs from that repo's root.
It relies on the standard sibling layout (all checked out next to each other):

```
<parent>/
  go-algorand/        <- you are here
  indexer/            <- indexer API spec lives here
  generator/          <- shared REST code generator (algorand/generator)
  go-algorand-sdk/
  java-algorand-sdk/
  js-algorand-sdk/
  py-algorand-sdk/
```

If any repo is elsewhere, ask the user for its path before proceeding. Use `-C`/`-jar`
paths rather than `cd` where practical, matching the repo's conventions.

## Arguments

An optional short description of what changed in go-algorand. If omitted, infer the
change from the current branch's diff against `master` and from the conversation. Also
use it (or the go-algorand branch name) to derive the `<slug>` for the SDK working
branches created in Step 2.

## How the pieces fit together (read first)

There are **two independent kinds** of propagation, and a given change may need one or
both. Classify the change before doing work:

1. **REST surface** — a change to the algod, indexer, or kmd HTTP API (new endpoint,
   new/changed request or response field, new model). This flows through the OpenAPI
   **specs** and the shared **`../generator`** tool.
2. **Shared serialized types** — a change to a Go type whose msgpack/JSON encoding must
   match across implementations (transaction fields, block structure, consensus
   params). For the **Go SDK only**, these are patched by
   `scripts/export_sdk_types.py`. The other SDKs pick these up through their own
   generated models (REST surface) or hand-written code.

Key facts about the generator:

- **Go, Java, and JS** REST clients are produced by `../generator` (a Java/Maven +
  Velocity tool). Its `scripts/generate_{go,java,typescript}.sh` default to the sibling
  layout above and read the **local** `algod.oas2.json` / `indexer.oas2.json`, so they
  reflect *unmerged* go-algorand changes.
- **Python has no generator** — `py-algorand-sdk` clients and models are hand-written.
- Normally a **nightly GitHub Action** in each of the Go/Java/JS SDKs regenerates from
  the upstream specs and opens a PR against the SDK's default branch. Running the
  generator locally is how you preview or land a change before it merges. Say so to the
  user if the change is already merged upstream and they might prefer the nightly PR.
- The **oas2 (Swagger 2.0) JSON is the source of truth**. `algod.oas2.json` is
  hand-edited; if the change adds/alters an endpoint, that edit must already exist in
  `go-algorand` (and, for indexer endpoints, in `../indexer/api/indexer.oas2.json`).
  This skill regenerates the *derived* artifacts, it does not invent endpoint specs.

## Step 1: Classify the change

From the branch diff, decide which paths were touched:

```bash
git diff --stat master... | cat
```

- REST/algod: `daemon/algod/api/algod.oas2.json`, `daemon/algod/api/server/v2/handlers.go`.
- REST/kmd: `daemon/kmd/...`.
- REST/indexer: lives in `../indexer` (spec `../indexer/api/indexer.oas2.json`).
- Shared types: `data/transactions/...`, `data/basics/...`, `data/bookkeeping/block.go`,
  `config/consensus.go`, `protocol/...` — cross-reference the export list in
  `scripts/export_sdk_types.py` to see which of these it actually mirrors.

State to the user which SDK-facing surfaces are in play; that set is exactly the SDKs to
preflight in Step 2.

## Step 2: Preflight — every target SDK exists, is current, and on a fresh branch

The generator and the type-export script **write into** the SDK trees, and the generator
is destructive (`find … -delete`). So before any generation, guarantee a clean, current
starting point for each SDK the change touches (and `../indexer` if an indexer spec is
involved). Do **not** reuse an existing feature branch — always branch fresh from
`upstream`.

The sibling layout uses `upstream` = `algorand/<repo>` and `origin` = the user's fork.
The four SDKs and `indexer` currently default to the **`main`** branch (go-algorand and
generator use `master`), but this has changed before — always confirm per repo rather
than assuming, by querying the remote's HEAD:

```bash
git -C <repo> remote show upstream | sed -n 's/.*HEAD branch: //p'   # e.g. main
```

For **each** target SDK (`go-algorand-sdk`, `java-algorand-sdk`, `js-algorand-sdk`,
`py-algorand-sdk`):

1. **Exists?** If the directory is missing, stop and tell the user — do **not** auto-clone,
   because `origin` should be their fork, not `upstream`. Ask them to clone it (or give
   its path if non-standard).
2. **Clean?** `git -C <sdk> status --porcelain` must be empty. If not, stop and ask the
   user to stash/commit/discard — the generator would destroy uncommitted work.
3. **Fetch upstream and branch fresh from its default branch:**
   ```bash
   DEF=$(git -C <sdk> remote show upstream | sed -n 's/.*HEAD branch: //p')   # usually main
   git -C <sdk> fetch upstream
   git -C <sdk> checkout -b sdk-update-<slug> "upstream/$DEF"
   ```

Use one `<slug>` (from the argument or the go-algorand branch) for every SDK so the
branches line up. Report the created branches to the user.

Note: `../generator` (Step 4) is a plain checkout that may carry local template tweaks —
do **not** reset it. And `../go-algorand` stays on the branch that carries the change
(Step 1); it is never reset to upstream.

## Step 3: Regenerate the go-algorand (and indexer) API specs

Only run the ones relevant to the change. These regenerate the derived specs/served
code that the generator and the SDKs consume.

```bash
# kmd swagger (generated FROM Go source; needs libsodium built)
make rebuild_kmd_swagger

# algod: oas2.json is source of truth; touch forces the oas3 + server regen
touch daemon/algod/api/algod.oas2.json
make -C daemon/algod/api generate

# indexer (only if an indexer endpoint/type changed)
make -C ../indexer/api generate
```

Sanity-check the result: `git -C . diff --stat daemon/` and, if touched,
`git -C ../indexer diff --stat`. The algod oas3 step calls out to a swagger converter
service — if it fails offline, tell the user rather than pressing on with a stale spec.

## Step 4: Ensure the generator is available

```bash
GEN=../generator
[ -d "$GEN" ] || git clone git@github.com:algorand/generator.git "$GEN"
```

The generator is a Maven project. Its wrapper scripts run `mvn package` unless given
`-s`/`--skip-build`. Build it **once** this session (drop `-s` on the first
`generate_*.sh` call); reuse the jar with `-s` on subsequent calls. If
`../generator/target/generator-*-jar-with-dependencies.jar` already exists and the
generator repo is unchanged, `-s` is safe from the start.

The generator is destructive — Step 2 already put each SDK on a clean branch from
upstream, so it is safe to run. Still verify you are generating from the **go-algorand
branch that carries the change** (`git branch --show-current` here): the generator
mirrors the *local* spec exactly, so generating from the wrong base silently
adds/removes models.

## Step 5: Go SDK (`../go-algorand-sdk`)

Two parts — do both if the change spans both surfaces.

**Shared types** (run from the go-algorand root; the script targets `../go-algorand-sdk`):

```bash
python3 scripts/export_sdk_types.py
```

This extracts the exported types/vars/funcs and `gofmt`s them into the SDK's `types/`,
`protocol/`, and `protocol/config/`. If the change added a brand-new type that should
be mirrored, add an `export_type(...)` line to the script's `__main__` block first (see
the existing entries), then re-run.

**REST clients** (run from the generator repo; defaults to the sibling specs/SDK):

```bash
../generator/scripts/generate_go.sh          # first call: builds the jar
```

This deletes and regenerates `client/v2/{algod,indexer,common/models}` (preserving the
handful of hand-written files the script `-name`-excludes) and runs `go fmt`.

**Removals need manual cleanup.** The generator deletes generated models that are no
longer in the spec, but *preserves* the hand-written wrappers on its `-name` exclusion
list. If the go-algorand change **removed** an endpoint, a preserved wrapper can be left
referencing a now-deleted model and the build fails (e.g. removing dryrun deletes
`models/dryrun_response.go` but keeps hand-written `client/v2/algod/dryrun.go`, which
then won't compile). When the compile check below fails with `undefined: models.X`,
manually delete the corresponding preserved hand-written file(s) and re-check. The same
applies to Java/JS (their preserve lists are `LedgerStateDelta.java`, and the
hand-written request classes / kmd client respectively).

**Verify:**

```bash
make -C ../go-algorand-sdk build   # go generate ./logic + compile check
make -C ../go-algorand-sdk lint    # client/v2 is lint-excluded
make -C ../go-algorand-sdk unit    # unit + cucumber unit tags
```

## Step 6: Java SDK (`../java-algorand-sdk`)

```bash
../generator/scripts/generate_java.sh -s     # -s: reuse jar built in Step 5
```

Regenerates `src/main/java/com/algorand/algosdk/v2/client/{model,algod,indexer}`
(preserving `LedgerStateDelta.java`). This is a Maven project with no `make build`
target — compile-check with `mvn -f ../java-algorand-sdk/pom.xml -q compile`; its
Makefile test targets are `make -C ../java-algorand-sdk unit` / `integration`.

Note: the Java **kmd** client (`.../kmd/client/`) is legacy swagger-codegen and is *not*
regenerated here — a kmd API change needs a manual update (rare; flag it).

## Step 7: JavaScript SDK (`../js-algorand-sdk`)

```bash
../generator/scripts/generate_typescript.sh -s
```

Only the `src/client/v2/{algod,indexer}/models/types.ts` files are generated (the script
runs `npm run format` afterward). **Per-endpoint request classes and the entire kmd
client are hand-written.** So if the change adds a *new endpoint*, the generated models
update automatically but a human must add the matching request class by hand (mirror an
existing sibling file under `src/client/v2/algod/` or `.../indexer/`). Point this out
and offer to draft it. Then build/test (`npm ci && npm run build`, per its README).

## Step 8: Python SDK (`../py-algorand-sdk`) — manual

No generator. Edit by hand, guided by the go-algorand diff:

- New/changed endpoint → edit `algosdk/v2client/algod.py` or `indexer.py` (thin
  `requests` wrappers; mirror an adjacent method).
- New/changed model → add or update `algosdk/v2client/models/*.py` following the
  existing `openapi_types`/`attribute_map` class shape.
- Then `make -C ../py-algorand-sdk lint` and `make -C ../py-algorand-sdk generate-init`
  (the latter only refreshes `algosdk/__init__.pyi` stubs), plus its test target.

## Step 9: Review

For each SDK actually touched:

```bash
git -C ../go-algorand-sdk  diff --stat | cat
git -C ../java-algorand-sdk diff --stat | cat
git -C ../js-algorand-sdk   diff --stat | cat
git -C ../py-algorand-sdk   diff --stat | cat
```

Confirm the diff matches the intended change and nothing unrelated regenerated (a stale
spec or a generator version bump can churn unrelated files — if so, stop and
investigate). **Scan for deletions specifically** (`git -C <sdk> status --porcelain | grep '^ D'`):
unexpected `D` entries mean the local spec lacks something the SDK still has — either
the go-algorand base is wrong, or the change genuinely removed it and a preserved
hand-written wrapper now needs deleting too (see Step 5). Summarize per-SDK what changed
and call out anything left manual (Python edits, JS request classes, kmd, wrapper
deletions).

## Notes

- Do not commit or open PRs in any repo unless the user asks; leave working trees on the
  Step 2 branches for review.
- If the go-algorand change is already merged upstream, remind the user the nightly
  codegen Actions will regenerate Go/Java/JS automatically — running locally is for
  previewing or for unmerged work.
- `scripts/export_sdk_types.py` deliberately does **not** export a few types (e.g.
  `BoxRef`) that were reshaped in the SDK; don't "fix" those by adding them.
- A new consensus parameter usually also needs a spec update — see the
  `spec-consensus-param` skill.
