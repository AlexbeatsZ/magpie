# Magpie Windows fork

Portable Windows x64 desktop build for managing agent resources and sessions.

- Agent inventory builds resource descriptors directly. Automatic model wiring,
  native integration plans and their obsolete tests have been removed.
- Library continues to distribute MCP, instructions and Skills through its
  existing adapters. Model endpoints and credentials remain manually configured.
- Native session references use `device:harness:native-id`, shared with
  agent-history. Native IDs and existing stored session keys remain unchanged.
- Legacy profile documents remain readable in backups. Agent model fields are
  skipped during restore and sync. Uninstall does not migrate agent sessions,
  edit agent configuration or restart client processes.
- Release builds contain the application, LICENSE and THIRD_PARTY_NOTICES only.
  Data, logs and WebView2 profiles are created beside the executable when run.
  The repository now has its own build/publish workflow and update manifests.

Download `magpie-windows-amd64.zip` and extract it to a writable folder. Run
`magpie.exe`. The separate EXE asset is provided for the existing updater format.
Do not copy a new empty data folder over existing data.

This release is a preview. The fork's resource/identity/manual-model checks,
Windows and Linux compilation and static checks are the release validation
scope. The complete upstream Windows suite still contains legacy behavior tests
and platform/environment failures (plugins, system proxies and terminal tests);
it is not green. Real-client UI acceptance and other-platform desktop packages
are not claimed. GitHub's stable latest feed remains unchanged by a prerelease.
