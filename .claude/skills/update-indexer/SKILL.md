---
name: update-indexer
description: Propagate a new/changed go-algorand block-header or transaction field into the indexer (and conduit) data-ingestion stack — SDK dependency bump, Postgres encoding/schema, REST API spec + converters, and conduit pass-through/filter regen. Use when the user asks to "update indexer", "add a field to indexer", or "propagate a block/txn field to indexer and conduit".
argument-hint: [block-header or transaction field name]
allowed-tools: [Bash, Read, Edit, Write, Grep, Glob]
---

# Update indexer (and conduit) for a new block-header or transaction field

Propagate a block-header or transaction field that was added/changed in `go-algorand`
into the indexing stack so the field is stored and queryable. Sibling layout:

```
<parent>/
  go-algorand/        <- you are here (source of the field)
  go-algorand-sdk/    <- indexer & conduit consume the field via this SDK
  indexer/            <- Postgres storage + REST API
  conduit/            <- block ingestion pipeline that feeds indexer
```

## Prerequisite: the field must already be in the Go SDK

Both indexer and conduit import **only** `github.com/algorand/go-algorand-sdk/v2` — they
do **not** import go-algorand directly. Blocks and transactions are stored and passed as
the SDK's `types.BlockHeader` / `types.SignedTxnInBlock` (as JSON blobs in indexer's
case). Nothing here works until the field exists on those SDK types.

So **run `/update-sdks` first** (at least its Go-SDK path: `export_sdk_types.py` for
shared types, or the generator for REST models). If the field is tied to a **new
consensus version or consensus parameter**, the SDK's `protocol/config` must carry it
too (that is what `export_sdk_types.py` mirrors): indexer reads `config.Consensus[...]`
from the SDK, and `AddBlock` errors on a block whose `CurrentProtocol` is unknown to the
SDK's `config.Consensus`. Confirm the field is present in `../go-algorand-sdk/types`
before proceeding.

Dependency direction: **SDK → indexer → conduit** (conduit also imports
`github.com/algorand/indexer/v3`). Do indexer before conduit.

## Step 1: Preflight — repos exist, current, on a fresh branch

For `indexer` and `conduit` (only the ones the change touches):

1. **Exists?** If missing, stop and ask the user for its path — do not auto-clone.
2. **Clean?** `git -C <repo> status --porcelain` must be empty; otherwise stop and ask.
3. **Fresh branch from the algorand remote's default.** Remotes differ: indexer usually
   has `upstream` = `algorand/indexer` (with `origin` = a fork); conduit usually has just
   `origin` = `algorand/conduit`. Pick whichever remote points at `algorand/<repo>`,
   then derive its default branch dynamically (it has been both `develop` and `main`):
   ```bash
   REM=upstream; git -C <repo> remote get-url upstream >/dev/null 2>&1 || REM=origin
   DEF=$(git -C <repo> remote show "$REM" | sed -n 's/.*HEAD branch: //p')
   git -C <repo> fetch "$REM"
   git -C <repo> checkout -b indexer-update-<slug> "$REM/$DEF"
   ```

## Step 2: Point indexer (and conduit) at the SDK that has the field

The pinned SDK version in `go.mod` won't have an unmerged field. For local iteration,
add a temporary `replace` to the sibling checkout; finalize to a real released version
only once the SDK change is merged and tagged.

```bash
# indexer -> local SDK
go -C ../indexer mod edit -replace github.com/algorand/go-algorand-sdk/v2=../go-algorand-sdk
go -C ../indexer mod tidy
```

For conduit, replace **both** the SDK and indexer (since conduit consumes indexer/v3, and
you likely have unmerged indexer changes too):

```bash
go -C ../conduit mod edit \
  -replace github.com/algorand/go-algorand-sdk/v2=../go-algorand-sdk \
  -replace github.com/algorand/indexer/v3=../indexer
go -C ../conduit mod tidy
```

Flag to the user that these `replace` directives are temporary and must be removed (and
swapped for a version bump) before the change is committed/merged.

**Expect unrelated breakage if the SDK jump is large.** `go mod tidy` will succeed, but
`go build ./...` can then fail on breaking changes the SDK made *besides* your field —
e.g. scalar fields becoming named types (`NextProtocolApprovals` going from `uint64` to
`types.Round`, `ApplicationID` to `types.AppIndex`), which break existing callsites in
`api/converter_utils.go` and `api/handlers.go`. Prefer replacing/bumping to the
*minimal* SDK revision that carries your field. If unrelated compile errors appear,
reconcile them (they are real API changes to absorb), don't work around them — and tell
the user, since it widens the change well beyond the one field.

## Step 3: Classify the field — how much indexer work is needed

Indexer stores the whole header/txn as a JSON blob (`block_header.header jsonb`,
`txn.txn jsonb`) plus a few extracted index columns, so the amount of work depends on
the field's nature:

- **Tier A — a plain scalar/int/bool field.** Rides in the blob automatically via the
  embedded SDK type. **No storage, encoding, or migration change.** Only Step 5 (API
  exposure) is needed to surface it.
- **Tier B — an address, address-slice, or raw byte-string** that must render
  human-readably (base32/base64) in JSON. Also needs Step 4a (an encoding override).
- **Tier C — a field that must be independently indexed/searchable, or a brand-new
  transaction type.** Also needs Step 4b (schema/migration/write-path) and search wiring.

State which tier applies before editing.

## Step 4: Indexer storage layer (skip for Tier A)

**4a. Encoding override (Tier B).** In
`../indexer/idb/postgres/internal/encoding/types.go`, add a `...Override` field (with the
explicit `codec:"..."` tag and readable type) to the wrapper struct that embeds the SDK
type — `blockHeader` for header fields, `transaction`/`signedTxnWithAD`/`evalDelta` for
txn fields. Then add matching lines in the paired `convert*`/`unconvert*` functions in
`../indexer/idb/postgres/internal/encoding/encoding.go` (e.g. `convertBlockHeader` /
`unconvertBlockHeader`). Model it on the existing `Proposer`/address overrides. Note the
codec uses `ErrorIfNoField = true`, so encode and decode must stay symmetric.

**4b. Indexed column / new txn type (Tier C).**
- Schema: edit `../indexer/idb/postgres/internal/schema/setup_postgres.sql`, then
  regenerate the compiled Go: `go -C ../indexer generate ./idb/postgres/internal/schema/`
  (produces `setup_postgres_sql.go`; also covered by top-level `make`).
- Migration: append a new `{migrationFunc, blocking, "desc"}` entry to the **end** of the
  `migrations` slice in `../indexer/idb/postgres/postgres_migrations.go` (never reorder
  or remove existing entries) and implement `migrationFunc` (use the `sqlMigration`
  helper for pure DDL).
- Write path: update `../indexer/idb/postgres/internal/writer/write_txn.go` only if the
  field feeds an extracted column (`transactionAssetID`) or a new `typeenum`
  (`idb/txn_type_enum.go`, `idb.GetTypeEnum`).

## Step 5: Indexer REST API

1. Add the field to the right definition in `../indexer/api/indexer.oas2.json` (the
   `Block` schema for header fields; the relevant transaction sub-schema —
   `TransactionPayment`, `TransactionKeyreg`, `TransactionApplication`, etc. — for txn
   fields). This oas2 JSON is the source of truth.
2. Regenerate: `make -C ../indexer/api generate` (converts to `indexer.oas3.yml` via the
   swagger converter, then runs `oapi-codegen` to rebuild `api/generated/v2/types.go` and
   `routes.go`). Two environment gotchas: the converter step needs network
   (`converter.swagger.io`) — if offline, tell the user; and the Makefile installs
   `oapi-codegen` with `go install` then calls the bare binary, so
   **`$(go env GOPATH)/bin` must be on `PATH`** or it fails with
   `oapi-codegen: No such file or directory` (`export PATH="$(go env GOPATH)/bin:$PATH"`).
3. Wire the DB value into the generated response struct in
   `../indexer/api/converter_utils.go`:
   - Header → `hdrRowToBlock` (use a helper from `api/pointer_utils.go`, e.g.
     `uint64PtrOrNil`, `addrPtr`, `byteSliceOmitZeroPtr`).
   - Transaction → `signedTxnWithAdToTransaction`, in the correct `switch stxn.Txn.Type`
     branch / sub-struct.
4. **Searchable?** Only if the user wants to filter/query by the field: add parsing in
   `transactionParamsToTransactionFilter` / `blockParamsToBlockFilter`, extend the filter
   struct in `../indexer/idb/idb.go`, and add SQL in the `postgres.go` query builders.

## Step 6: Build & test indexer

```bash
make -C ../indexer            # builds binary + regenerates schema Go + idb mocks
make -C ../indexer test
make -C ../indexer lint
```

If you changed the `idb.IndexerDb` interface (Tier C search), `make` regenerates
`idb/mocks/IndexerDb.go` via mockery — commit that too.

## Step 7: Conduit

Conduit is mostly a **pass-through** of SDK types (`data.BlockData` holds
`sdk.BlockHeader` and `[]sdk.SignedTxnInBlock`), so a new field flows through with no
code change once Step 2's SDK bump is in place. Two conditional tasks:

- **Filterable transaction field:** if users should be able to filter on it, regenerate
  the reflection map:
  ```bash
  go -C ../conduit generate ./conduit/plugins/processors/filterprocessor/fields/
  ```
  This rebuilds `generated_signed_txn_map.go` (and `Filter_tags.md`) from the updated SDK
  `SignedTxnWithAD`. Block-header fields are not filterable here, and nothing breaks if a
  new field simply isn't added to the map.
- Build/test: `make -C ../conduit` (build) and `make -C ../conduit test` / `lint`.

## Step 8: Review

```bash
git -C ../indexer diff --stat | cat
git -C ../conduit diff --stat | cat
```

Confirm the diff matches the field and tier. **Re-check the `go.mod` files**: the
temporary `replace` directives from Step 2 are for local testing only — before
committing, remove them and bump `go-algorand-sdk/v2` (and, in conduit, `indexer/v3`) to
a real released version once the upstream changes are merged/tagged. Summarize per repo
what changed and what remains manual (a released-version bump, tests, any search wiring
you deferred).

## Notes

- Do not commit or open PRs unless asked; leave working trees on the Step 1 branches.
- Tier A is the common case — a plain field needs only the SDK bump plus the API spec +
  converter edit; conduit needs nothing. Don't add migrations or encoding overrides
  "just in case."
- A new consensus parameter also needs a go-algorand spec update — see the
  `spec-consensus-param` skill — and the SDK sync in `/update-sdks`.
