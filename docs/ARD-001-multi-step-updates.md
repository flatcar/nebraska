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
Syncers mirror an upstream Nebraska. They receive updates the same way as regular clients - **one flagged package per response** - and walk floors one at a time:
- The upstream returns the lowest floor above the syncer's reported version (`is_floor="true"`), then the target (`is_target="true"`) once all floors have been passed. A package can carry both flags.
- The syncer keeps a **walk cursor** (the version it reports upstream) separate from its channel: an intermediate floor advances only the cursor and is recorded locally, while the **channel advances directly to the target** and never points at an intermediate floor.
- **Floor-capable syncers** advertise `multi_manifest_ok=true`. A floor-unaware syncer that omits it is **blocked with `NoUpdate`** when a floor lies ahead, so it cannot walk past a floor it cannot record.
- Syncers are identified by `InstallSource="scheduler"` in the Omaha request.
- Every syncer request is gated by the upstream group's rollout policy (updates enabled, office hours, and the group's update limits), so a mirror honors the same policy as regular clients. No update grant is recorded and rollout state is left untouched: a syncer is a fake, stats-excluded instance, so a grant would be cosmetic. The walk is not throttled by the group's per-period/concurrency limits, because a fake instance never counts toward those totals.
- During a Nebraska-to-Nebraska upgrade, upgrade downstream syncers before upstreams. `multi_manifest_ok` remains the floor-safety gate (a syncer that does not advertise it is blocked when a floor lies ahead), but a downstream on a version that predates the single-package walk will advertise it and still mishandle a floor, so upgrade order is what guarantees safety.

### Safety Rules & Constraints

#### Universal Constraints
1. **Floor/Blacklist Mutual Exclusion**: Packages cannot be both floor AND blacklisted for same channel
2. **Channel Target Protection**: Channel's current package cannot be blacklisted
3. **Cross-channel Independence**: Package can be floor for one channel and blacklisted for another

#### Syncer-specific Constraints
1. **Ordered floor recording**: a floor is recorded before the walk cursor advances, so a failure retries the same floor and never skips it
2. **Policy without grant**: every request is gated by the group's rollout policy, but no update grant or rollout-state change is recorded for the fake syncer instance
3. **Package Verification**: existing packages verified for hash/size match before reuse
4. **Download Cleanup**: failed downloads cleaned up to prevent orphaned files
5. **Legacy Syncer Safety**: floor-unaware syncers (no `multi_manifest_ok`) blocked when a floor lies ahead

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
- Legacy (floor-unaware) syncers blocked when floors are present (requires upgrade)
- Additional database queries for floor checking
- A syncer converges on a multi-floor target over several sync cycles (one package per cycle) rather than in a single response

## Dependencies

### go-omaha Library
Enhanced go-omaha library with:
- Floor attributes on the manifest (`IsFloor`, `FloorReason`, `IsTarget`)
- `MultiManifestOK` capability flag, used as the syncer's "floor-aware" signal

## References
- [Flatcar Discussion #1831](https://github.com/flatcar/Flatcar/discussions/1831)