# ARD-001: Multi-Step Updates (Floor Packages)

## Status
Implemented

## Context
Flatcar requires mandatory intermediate versions (floors) when updating across major versions to prevent compatibility issues and system failures.

## Decision

### Core Architecture
- **Channel-specific floors** stored in new `channel_package_floors` junction table with indexes
- **Dynamic metadata**: `IsFloor` and `FloorReason` populated at query time, not stored in package table
- **Reuse existing tracking**: Utilize existing `last_update_version` field for already-granted state management
- **Atomic package creation**: New `AddPackageWithMetadata` API for complete metadata insertion

### Update Behavior

#### Regular Clients
- Receive single package (next floor or target) and progress sequentially through floors
- Already-granted instances with `last_update_version` can continue progression without re-evaluation
- NULL `last_update_version` triggers completion to force fresh grant cycle

#### Syncers (Nebraska instances)
- **Modern syncers with `multi_package_ok=true`**: Receive all packages (floors + target) in one response
- **Legacy syncers without `multi_package_ok`**: Blocked with `NoUpdate` response when floors exist
- Syncers identified by `InstallSource="scheduler"` in Omaha request

#### Target Detection (Syncer-specific)
For multi-manifest responses, syncers use this priority:
1. **Explicit**: Manifest with `is_target="true"` attribute
2. **Implicit**: Last manifest that is NOT a floor (backward compatibility)
3. **None**: All manifests are floors (valid - no channel update)

### Safety Rules & Constraints

#### Universal Constraints
1. **Floor/Blacklist Mutual Exclusion**: Packages cannot be both floor AND blacklisted for same channel
2. **Channel Target Protection**: Channel's current package cannot be blacklisted
3. **Cross-channel Independence**: Package can be floor for one channel and blacklisted for another

#### Syncer-specific Constraints
1. **Atomic Floor Operations**: Floor marking failures abort entire update
2. **Package Verification**: Existing packages verified for hash/size match before reuse
3. **Download Cleanup**: Failed downloads cleaned up to prevent orphaned files
4. **Legacy Syncer Safety**: Syncers without multi-package support blocked when floors exist

### API Endpoints

#### Floor Management
- `POST /api/v1/apps/{app_id}/channels/{channel_id}/packages/{package_id}/floor` - Mark as floor
- `DELETE /api/v1/apps/{app_id}/channels/{channel_id}/packages/{package_id}/floor` - Unmark as floor
- `GET /api/v1/apps/{app_id}/channels/{channel_id}/packages/floors` - List floor packages

#### Package Response Fields
- `is_floor`: Boolean indicating floor status
- `floor_reason`: Text explanation for floor requirement

## Consequences

### Positive
- Safe update paths preventing incompatible version jumps
- Backward compatible with existing single-step updates
- Sequential progression ensures system stability
- Channel-specific flexibility for different update strategies
- Atomic operations prevent partial states

### Negative
- Legacy syncers blocked when floors present (requires upgrade)
- Additional database queries for floor checking
- Increased complexity in update decision logic

## Dependencies

### go-omaha Library
Enhanced go-omaha library with:
- Multi-manifest support (`Manifests` array replacing single `Manifest`)
- Floor attributes (`IsFloor`, `FloorReason`, `IsTarget`)
- `MultiPackageOK` capability flag for syncers

## References
- [Flatcar Discussion #1831](https://github.com/flatcar/Flatcar/discussions/1831)