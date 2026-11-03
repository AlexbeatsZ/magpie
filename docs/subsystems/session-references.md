# Session references and agent-history

Magpie uses agent-history's `device:harness:native-id` reference format. The
native session ID remains the source of truth. No session files, resume IDs,
gateway attribution keys, or existing `agent:id` storage keys are migrated.

## Ownership and compatibility

[`sessions.Reference`](../../internal/sessions/reference.go) creates references
for native sessions on this computer. The existing three-device history bridge
names Windows computers containing `rog` as `rog`, other Windows computers as
`omen`, and macOS as `mac`. Other Linux hosts use their short hostname.
Magpie's `agy` agent is named `antigravity` in the history bridge; other harness
names are preserved. A reference is an identity, not a claim that deja-vu has
an adapter for that agent or has already indexed it.

Gateway-only sessions have no confirmed source computer, and WSL transcripts
belong to a separate environment. Neither receives a local history reference.
The gateway's existing IDs and session views are unchanged.

## Lookup and UI

The session list, management list and detail API add `reference` next to the
existing native `id`. The detail's `GET /api/sessions/one?key=...` also accepts a
full reference, `[session device:harness:prefix]`, or an unqualified native ID
or prefix of at least four characters. Old exact `agent:id` queries still work.

[`sessions.ResolveReference`](../../internal/sessions/reference.go) uses the
existing session readers and their consent gate. It searches the complete
agent listing, accepts a unique prefix, prefers an exact ID over longer IDs,
and rejects ambiguity (HTTP 409). A foreign device reference is HTTP 400 and
names the corresponding `agent-history-<device>` server. Missing local sessions
are HTTP 404. No remote request or transcript replication is performed.

The Sessions page shows the native ID and a separately copyable session
reference. Pasting a reference into Search sessions finds and opens that
session, switching to its agent. Reference lookup uses the detail API and
does not search transcript text. A new query cancels an outstanding lookup.

Magpie-generated full references can be passed to agent-history's existing
`get_session({id: reference})`; its unique short prefixes and SessionStart hook
references can be pasted into Magpie. Neither the bridge nor its indexes need
to be replaced for this compatibility.

## Verification

Tests use native-format transcript fixtures, isolated homes and mocked browser
APIs. No live agent configuration, index, credential or running app is changed.

```text
go test -tags nogui ./internal/sessions ./internal/gui -run TestHistoryReference
node --test internal/gui/tests/resource-management.test.cjs
```
