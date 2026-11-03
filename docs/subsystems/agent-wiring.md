# Agent resources and manual model setup

Magpie manages agent resources through the Library and reads native sessions.
Each agent's model, endpoint and credentials remain under the user's control.

## Owners and runtime paths

| Responsibility | Owner |
| --- | --- |
| Agent names, icons, config locations and installation detection | `internal/agent`, `Inventory`, `InventoryDetected`, `Find` |
| Request attribution by client ID or User-Agent | `agent.Clients`, `usage.Agents` |
| MCP, instructions, skills, hooks and subagents | `internal/library`, its adapters and existing sync/change owner |
| Native sessions and portable references | `internal/sessions`; [session references](session-references.md) |
| Model/provider routing inside Magpie's gateway | `internal/provider`, `internal/gateway`; no agent config writes |
| Legacy profile documents in backups | `internal/profile`: JSON codecs and storage only |

Registry constructors build resource metadata directly. They do not create model
fields, native integration plans, launch commands or model writer closures.
`Inventory` returns this registry, including plugin descriptors and WSL targets;
it does not construct writable agents and strip them afterwards. Stopped WSL
distros remain unread until explicitly opened through the existing Library owner.

The Agents page links to Library and Sessions. Its state contains identity,
location and resource capabilities. The empty `fields` and `profiles` arrays are
kept for existing page/state readers. The old agent model picker, Profile actions
and model-writing HTTP routes are unavailable. CLI and TUI model setters are
removed too.

The Library is the only resource writer. Startup and linked-resource sync call
its existing owner; there is no second distribution or model-sync loop. See
[library.md](library.md) for ownership and change tracking.

## Compatibility boundaries

Some internal read interfaces still describe the models a gateway client asks
for, and old operation signatures remain as guards for callers in the fork.
`All` and `Detected` expose no model-managed agents. `Apply`, `Pick`, `Connect`,
`Reapply`, `Disconnect` and tuning operations reject writes with
`ErrWiringRemoved`. Catalog sync and rewire entry points have no agent effects.

`internal/profile` preserves older `profiles.json` and backup documents without
applying their model fields. New backups omit agent model settings. Restoring an
older backup reports its agent settings as skipped; resource data still goes
through the Library. Sync preserves legacy documents but does not read, merge or
restore local agent model choices.

Uninstall removes Magpie's own system integrations and files. It does not invoke
legacy agent disconnection, migrate Codex rollouts or restart an agent process.
The existing Library retention rules still protect resources agents use.

## Verification

- `go test -tags nogui ./internal/agent ./internal/profile ./internal/uninstall`
- `go test -tags nogui ./internal/gui -run 'Test(GatewayOnly|ResourceManagement|HistoryReference)'`
- `go test -tags nogui ./internal/backup ./internal/davsync -run 'Test(Legacy|SyncPreservesManual)'`
- `node --test internal/gui/tests/resource-management.test.cjs`

Tests use isolated homes and fake applications. Live agent configuration and
running client processes are not release-test fixtures.
