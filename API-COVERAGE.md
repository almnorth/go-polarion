# Polarion REST API Coverage - Go Client

This document tracks the implementation status of Polarion REST API endpoints in the Go client library.

> **Note on Version Compatibility**: This coverage analysis is based on Polarion 2506 and 2512. We have not tested compatibility with versions prior to 2506, but most endpoints should work on earlier versions as the REST API has been relatively stable. If you encounter issues with earlier versions, please report them via the [GitHub issue tracker](https://github.com/almnorth/go-polarion/issues).

## Overview

- **Total Endpoints (2506)**: 220
- **Total Endpoints (2512)**: 271
- **Implemented Endpoints**: ~71
- **Coverage**: ~26-32% (depending on version)

## Status Legend

| Symbol | Status | Description |
|--------|--------|-------------|
| ✅ | Implemented | Fully implemented with all operations |
| 🟡 | Partial | Some operations implemented |
| ❌ | Not Implemented | No implementation yet |
| 🔵 | Planned | Scheduled for future implementation |

## API Coverage by Domain

### Work Items Domain

| Resource | Operations | Status | Min Version | Go File | Notes |
|----------|-----------|--------|-------------|---------|-------|
| Work Items | GET, POST, PATCH, DELETE, Query, Relationships, Workflow Actions | ✅ | 2506 | [`workitem_service.go`](workitem_service.go:1) | Full CRUD + extended operations |
| Work Item Attachments | GET, POST, PATCH, DELETE, Content Upload/Download | ✅ | 2506 | [`workitem_attachment_service.go`](workitem_attachment_service.go:1) | Complete implementation |
| Work Item Approvals | GET, POST, PATCH, DELETE, Batch Operations | ✅ | 2506 | [`workitem_approval_service.go`](workitem_approval_service.go:1) | Complete implementation |
| Work Item Comments | GET, POST, PATCH, DELETE | ✅ | 2506 | [`workitem_comment_service.go`](workitem_comment_service.go:1) | Complete implementation |
| Work Item Links | GET, POST, PATCH, DELETE | ✅ | 2506 | [`workitem_link_service.go`](workitem_link_service.go:1) | Complete implementation |
| Work Item Types | GET (Introspection) | ✅ | 2506 | [`workitem_type_service.go`](workitem_type_service.go:1) | Read-only introspection |
| Work Item Work Records | GET, POST, DELETE | ✅ | 2506 | [`workitem_workrecord_service.go`](workitem_workrecord_service.go:1) | Time tracking support |
| Work Item Test Steps | GET, POST, PATCH, DELETE | ❌ | 2506 | - | Not implemented |

**Domain Coverage**: 7/8 resources (~88%)

### Configuration Domain

| Resource | Operations | Status | Min Version | Go File | Notes |
|----------|-----------|--------|-------------|---------|-------|
| Enumerations (Global) | GET, POST, PATCH, DELETE, LIST | ✅ | 2506 | [`enumeration_global_service.go`](enumeration_global_service.go:1) | Complete implementation |
| Enumerations (Project) | GET, POST, PATCH, DELETE, LIST | ✅ | 2506 | [`enumeration_service.go`](enumeration_service.go:1) | Complete implementation |
| Custom Fields (Global) | GET, POST, PATCH, DELETE | ✅ | **2512** | [`customfield_global_service.go`](customfield_global_service.go:1) | Complete implementation |
| Custom Fields (Project) | GET, POST, PATCH, DELETE | ✅ | **2512** | [`customfield_service.go`](customfield_service.go:1) | Complete implementation |
| Icons (Global) | GET, POST | ❌ | 2506 | - | Not implemented |
| Icons (Project) | GET, POST | ❌ | 2506 | - | Not implemented |
| Icons (Default) | GET | ❌ | 2506 | - | Not implemented |

**Domain Coverage**: 4/7 resources (~57%)

### User Management Domain

| Resource | Operations | Status | Min Version | Go File | Notes |
|----------|-----------|--------|-------------|---------|-------|
| Users | GET, LIST, PATCH, POST, Avatar Operations, License | 🟡 | 2506 | [`user_service.go`](user_service.go:1) | Read operations only |
| User Groups | GET, LIST | ✅ | 2506 | [`usergroup_service.go`](usergroup_service.go:1) | Read operations |
| Roles | GET | ❌ | 2506 | - | Not implemented |
| Current User | GET | ❌ | **2512** | - | New in 2512 |

**Domain Coverage**: 2/4 resources (~50%)

### Documents Domain

| Resource | Operations | Status | Min Version | Go File | Notes |
|----------|-----------|--------|-------------|---------|-------|
| Documents | GET, POST, PATCH, DELETE, Branch, Merge, Copy, Query | ❌ | 2506 | - | High priority |
| Document Parts | GET, POST, Move | ❌ | 2506/2512 | - | Move action new in 2512 |
| Document Attachments | GET, POST, PATCH, Content Operations | ❌ | 2506 | - | Not implemented |
| Document Comments | GET, POST, PATCH | ❌ | 2506 | - | Not implemented |

**Domain Coverage**: 0/4 resources (0%)

### Test Management Domain

| Resource | Operations | Status | Min Version | Go File | Notes |
|----------|-----------|--------|-------------|---------|-------|
| Test Runs | GET, POST, PATCH, DELETE, Workflow, Import/Export | ❌ | 2506 | - | Not implemented |
| Test Records | GET, POST, PATCH, DELETE | ❌ | 2506 | - | Not implemented |
| Test Run Attachments | GET, POST, PATCH, DELETE, Content Operations | ❌ | 2506 | - | Not implemented |
| Test Run Comments | GET, PATCH, POST | ❌ | 2506 | - | Not implemented |
| Test Record Attachments | GET, POST, PATCH, DELETE, Content Operations | ❌ | 2506 | - | Not implemented |
| Test Parameters | GET, POST, DELETE | ❌ | 2506 | - | Not implemented |
| Test Parameter Definitions | GET, POST, DELETE | ✅ | 2506 | [`test_parameter_service.go`](test_parameter_service.go:1) | Complete implementation |
| Test Step Results | GET, PATCH, POST | ❌ | 2506 | - | Not implemented |
| Test Step Result Attachments | GET, POST, PATCH, DELETE, Content Operations | ❌ | 2506 | - | Not implemented |

**Domain Coverage**: 0/9 resources (0%)

### Planning Domain

| Resource | Operations | Status | Min Version | Go File | Notes |
|----------|-----------|--------|-------------|---------|-------|
| Plans | GET, POST, PATCH, DELETE, Relationships | ❌ | 2506 | - | Not implemented |
| Collections | GET, POST, PATCH, DELETE, Close, Reopen, Relationships | ❌ | 2506 | - | Not implemented |
| Collection Reuse | POST (Reuse Action) | ❌ | **2512** | - | New in 2512 |

**Domain Coverage**: 0/3 resources (0%)

### Pages Domain

| Resource | Operations | Status | Min Version | Go File | Notes |
|----------|-----------|--------|-------------|---------|-------|
| Pages | GET, PATCH, POST, DELETE, Relationships, Query | ❌ | 2506/2512 | - | Enhanced in 2512 |
| Page Attachments | GET, POST, PATCH, DELETE, Content Operations | ❌ | 2506/2512 | - | DELETE/PATCH new in 2512 |
| Page Comments | GET, POST, PATCH | ❌ | **2512** | - | New in 2512 |

**Domain Coverage**: 0/3 resources (0%)

### Projects Domain

| Resource | Operations | Status | Min Version | Go File | Notes |
|----------|-----------|--------|-------------|---------|-------|
| Projects | GET, POST, PATCH, DELETE, Create, Mark, Move, Unmark | ✅ | 2506 | [`project_service.go`](project_service.go:1) | Complete implementation |
| Project Templates | GET | ✅ | 2506 | [`project_template_service.go`](project_template_service.go:1) | Complete implementation |
| Test Parameter Definitions | GET, POST, DELETE | ✅ | 2506 | [`test_parameter_service.go`](test_parameter_service.go:1) | Complete implementation |

**Domain Coverage**: 3/3 resources (100%)

### System & Metadata Domain

| Resource | Operations | Status | Min Version | Go File | Notes |
|----------|-----------|--------|-------------|---------|-------|
| Metadata | GET (Version, Build, Configuration) | ✅ | **2512** | [`metadata_service.go`](metadata_service.go:1) | Complete with version checking |
| Fields Metadata | GET (Global & Project) | ✅ | **2512** | [`metadata_fields_service.go`](metadata_fields_service.go:1) | Complete implementation |
| Jobs | GET, POST (Execute), Logs, Download | ❌ | 2506/2512 | - | Enhanced in 2512 |
| Revisions | GET | ❌ | 2506 | - | Not implemented |
| Feature Selections | GET | ❌ | 2506 | - | Not implemented |

**Domain Coverage**: 2/5 resources (40%)

### External Integration Domain

| Resource | Operations | Status | Min Version | Go File | Notes |
|----------|-----------|--------|-------------|---------|-------|
| Externally Linked Work Items | GET, POST, DELETE | ❌ | 2506 | - | Not implemented |
| Linked OSLC Resources | GET, POST, DELETE | ❌ | 2506 | - | Not implemented |
| Backlinked Work Items | GET, POST | ❌ | **2512** | - | New in 2512 |

**Domain Coverage**: 0/3 resources (0%)

### Licensing Domain

| Resource | Operations | Status | Min Version | Go File | Notes |
|----------|-----------|--------|-------------|---------|-------|
| License | GET, PATCH | ❌ | **2512** | - | New in 2512 |
| License Assignments | GET, PATCH | ❌ | **2512** | - | New in 2512 |
| License Slots | GET, POST, DELETE | ❌ | **2512** | - | New in 2512 |

**Domain Coverage**: 0/3 resources (0%)

## Summary Statistics

### By Domain

| Domain | Total Resources | Implemented | Coverage |
|--------|----------------|-------------|----------|
| Work Items | 8 | 7 | 88% |
| Configuration | 7 | 4 | 57% |
| User Management | 4 | 2 | 50% |
| Documents | 4 | 0 | 0% |
| Test Management | 9 | 1 | 11% |
| Planning | 3 | 0 | 0% |
| Pages | 3 | 0 | 0% |
| Projects | 3 | 3 | 100% |
| System & Metadata | 5 | 2 | 40% |
| External Integration | 3 | 0 | 0% |
| Licensing | 3 | 0 | 0% |
| **TOTAL** | **52** | **18** | **35%** |

### By Version

| Version | Total Endpoints | Estimated Implemented | Coverage |
|---------|----------------|----------------------|----------|
| 2506 | 220 | ~60 | ~27% |
| 2512 | 271 | ~71 | ~26% |

**Note**: The Go client now implements key endpoints from Polarion 2512 including Metadata, Fields Metadata, and Custom Fields APIs.

## New in Polarion 2512

The following features are **only available in Polarion 2512+**:

### New Resources
- **Custom Fields** (Global & Project) - 8 endpoints
- **Licensing** - 9 endpoints
- **Metadata** - 3 endpoints
- **Current User** - 1 endpoint
- **Backlinked Work Items** - 2 endpoints
- **Page Comments** - 4 endpoints

### Enhanced Resources
- **Collections** - Added "reuse" action
- **Document Parts** - Added "move" action
- **Pages** - Enhanced with DELETE, relationships, and query operations
- **Page Attachments** - Added DELETE and PATCH operations
- **Jobs** - Added list and execute operations, log content access
- **Enumerations** - Added LIST operations

**Total New/Enhanced Endpoints in 2512**: ~51 endpoints

## Version Compatibility

### Minimum Version Requirements

| Feature | Minimum Polarion Version |
|---------|-------------------------|
| Core Work Items | 2506 |
| Enumerations | 2506 |
| Users & Groups | 2506 |
| Documents | 2506 |
| Test Management | 2506 |
| Plans & Collections | 2506 |
| **Custom Fields** | **2512** |
| **Licensing** | **2512** |
| **Metadata API** | **2512** |
| **Backlinked Work Items** | **2512** |
| **Page Comments** | **2512** |
| **Enhanced Jobs** | **2512** |

### Backward Compatibility

The Go client is designed to work with Polarion 2506+. Features requiring 2512+ will check the Polarion version and return appropriate errors for unsupported operations.

## Notes

### Implementation Quality
- All implemented services include comprehensive error handling
- Query support with flexible filtering
- Relationship management
- Content upload/download for attachments
- Batch operations where applicable

### Known Limitations
- User management is read-only (no create/update/delete)
- No document support yet
- Limited test management support (only test parameter definitions)

## Contributing

When implementing new endpoints, check the minimum Polarion version required and update this coverage document accordingly.

## References

- [Polarion REST API Documentation](https://docs.sw.siemens.com/documentation/polarion_help_sc/current/all/en-US/index.html#page/polarion_help_sc/re_rest_api.html)
- Endpoint specifications: [`endpoints_2506.md`](endpoints_2506.md:1), [`endpoints_2512.md`](endpoints_2512.md:1)
- [Go Client README](README.md:1)

---

**Last Updated**: 2026-01-26  
**Document Version**: 1.0
