# Changelog

All notable changes to `go-polarion` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Multi-enumeration custom field support.** Multi-value custom fields
  (multi-enumeration and multi-value string) can now be read and written as
  `[]string` instead of only through the raw custom fields map.
  - New accessors on `CustomFields`: `GetStringSlice` / `GetEnums` and
    `SetStringSlice` / `SetEnums`. Reading tolerates a JSON array, a `[]string`,
    or a bare string (yielding a single-element slice).
  - `LoadCustomFields` / `SaveCustomFields` now map `[]string`, `*[]string`, and
    slices of named string types (e.g. `[]Platform`); previously such fields
    failed with `field must be a pointer type`. A `nil` slice removes the field,
    an empty non-nil slice sends `[]` (clearing the field in Polarion).
  - `CustomFieldType.IsMulti()` reports the multi-value flag, accepting both the
    `multi` and `multiValue` JSON keys Polarion has used.
  - The code generator emits `[]string` for multi-value string and enumeration
    fields when the metadata exposes the flag.
- **All Go numeric widths for integer and float custom fields.**
  `LoadCustomFields` / `SaveCustomFields` previously supported only `*int` and
  `*float64` and failed with `unsupported field type: int64` on anything else.
  They now map `*int8`…`*int64`, `*uint`…`*uint64`, `*float32`, and named types
  with those kinds (e.g. `type ScopePosition int64`).
  - New `CustomFields.GetInt64` accessor, converting gracefully from any numeric
    representation: signed/unsigned integers, `float32`/`float64`, `json.Number`,
    and numeric strings. `GetInt` and `GetFloat` accept the same inputs.
  - Values that do not fit the target field (overflow, or a negative value in an
    unsigned field) now report an error instead of being silently truncated.
  - Integers are saved as `int64`, so values above 2^53 round-trip exactly
    instead of losing precision through `float64`.

### Fixed

- `LoadCustomFields` no longer panics when a custom field is declared with a named
  string, numeric, or boolean type (e.g. `*ScopePosition`); such fields were
  matched by kind but assigned as their underlying type.

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
