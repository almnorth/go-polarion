# Changelog

All notable changes to `go-polarion` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.17] - 2026-06-11

### Added

- **Native external work item link support.** New `WorkItemExternalLinks`
  service on `ProjectClient` exposing `Get`, `List`, `Create`, and `Delete`
  for work items linked from another Polarion repository or server.
  - New `WorkItemExternalLink` type (with `…Attributes`, `…Links`, and `…Meta`).
  - `NewWorkItemExternalLink(role, workItemURI)` constructor.
  - `ParseExternalLinkID` / `BuildExternalLinkID` helpers for the composite
    `{projectId}/{workItemId}/{roleId}/{hostname}/{targetProjectId}/{linkedWorkItemId}`
    link ID format.
  - `Create` and `Delete` batch multiple links in a single request.
- **Backlinked work item includes.** Work items can now resolve incoming
  ("backlink") relationships inline.
  - `WorkItem.BacklinkedWorkItemsInline` and
    `WorkItem.ExternallyLinkedWorkItemsInline` fields carry included resources.
  - New relationship fields on the work item resource:
    `BacklinkedWorkItems`, `ExternallyLinked`, plus the full relationship set
    (`Assignee`, `Author`, `Categories`, `LinkedWorkItems`, `Attachments`,
    `Comments`, `LinkedOslc`, `Module`, `ModuleFolder`, `Plan`, `Project`,
    `Votes`, `Watches`, `WorkRecords`, `ApprovalRecords`).
  - Backlink accessors on `WorkItemLink`: `GetBacklinkSourceWorkItemID(Short)`,
    `GetBacklinkSourceProjectID`, `GetBacklinkParentWorkItemID(Short)`,
    `GetBacklinkParentProjectID`, and the `GetPrimaryWorkItemID(Short)` /
    `GetPrimaryProjectID` family.
- **Sparse-field selectors for the new include types:**
  `FieldSelector.WithExternallyLinkedWorkItemFields` and
  `FieldSelector.WithBacklinkedWorkItemFields`.
- `WithGetInclude(...)` option to request included resources on single-item
  `Get` calls.

### Changed

- `workitem_service.go` now hydrates included external and backlinked work
  items from the JSON:API `included` section into the inline fields above.
- Expanded `docs/API-REFERENCE.md`, `README.md`, and `API-COVERAGE.md` to
  cover external links and backlinked includes.

## [0.1.16] - 2026-06-03

### Added

- Project-level query support.

## [0.1.5] - 2026-04-17

### Fixed

- False-positive change detection on HTML fields during comparison.

## [0.1.4] - 2026-03-20

### Added

- `include` parameter support.

### Fixed

- String trimming during comparison.

## [0.1.3] - 2026-02-09

### Changed

- Increased parallelism for concurrent requests.

## [0.1.2] - 2026-01-28

### Fixed

- Assorted bug fixes.

## [0.1.1] - 2026-01-27

### Added

- Release `0.1.1`.

## [0.1.0] - 2026-01-26

### Added

- Initial release.

[0.1.17]: https://github.com/almnorth/go-polarion/compare/v0.1.16...v0.1.17
[0.1.16]: https://github.com/almnorth/go-polarion/compare/v0.1.5...v0.1.16
[0.1.5]: https://github.com/almnorth/go-polarion/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/almnorth/go-polarion/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/almnorth/go-polarion/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/almnorth/go-polarion/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/almnorth/go-polarion/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/almnorth/go-polarion/releases/tag/v0.1.0
