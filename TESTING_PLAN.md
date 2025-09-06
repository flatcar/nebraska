# Manual Testing Plan - Multi-Step Updates with Floor Packages

## Test Environment Setup

### 1. Start Nebraska with Test Database

```bash
# Start everything (PostgreSQL + Nebraska with backend and frontend)
cd backend
docker compose -f docker-compose.test.yaml up -d
cd ..

# Access Nebraska at: http://localhost:8002
# PostgreSQL will be at: localhost:8001

# To stop everything:
# cd backend && docker compose -f docker-compose.test.yaml down

# To reset and restart with fresh database:
# cd backend && docker compose -f docker-compose.test.yaml down -v
# docker compose -f docker-compose.test.yaml up -d
```

### 2. Load Test Data

Copy and paste this entire command block to load test data:

```bash
docker exec -i backend_postgres_1 psql -U postgres nebraska_tests << 'EOF'
-- Note: The test database already has the application, channels, and groups created by migrations
-- We only need to add our test packages

-- Create test packages with REAL production update metadata
INSERT INTO package (id, type, version, url, filename, description, size, hash, application_id, arch)
VALUES
  ('89287967-c45c-4f7a-862c-68560b45486d', 1, '4081.3.5', 'https://update.release.flatcar-linux.net/amd64-usr/4081.3.5/', 
   'flatcar_production_update.gz', 'Flatcar 4081.3.5', '507727150', 
   'J2wtrup4KtSE6NZd2sqBeRgZQdw=', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 1),
  ('42a49084-171c-4801-a343-aaf91dd71d12', 1, '4230.2.0', 'https://update.release.flatcar-linux.net/amd64-usr/4230.2.0/', 
   'flatcar_production_update.gz', 'Flatcar 4230.2.0', '498780404', 
   'QLF8SIat97vSvD5fP4VYoBokLcc=', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 1),
  ('1358f02b-0b67-40b4-b6fa-a6e2fd4afa63', 1, '4230.2.1', 'https://update.release.flatcar-linux.net/amd64-usr/4230.2.1/', 
   'flatcar_production_update.gz', 'Flatcar 4230.2.1', '498407990', 
   '7PW61iO2lNWvdlo3yiDj0MZAVtk=', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 1),
  ('c9c275ba-d289-4537-9cdc-399cebdeaac1', 1, '4230.2.2', 'https://update.release.flatcar-linux.net/amd64-usr/4230.2.2/', 
   'flatcar_production_update.gz', 'Flatcar 4230.2.2', '499005872', 
   'eloy9wIKfw6RkxHgExeV5Asizck=', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 1);

-- Create flatcar actions with REAL SHA256 hashes in base64 format
INSERT INTO flatcar_action (sha256, package_id, event, is_delta, disable_payload_backoff)
VALUES
  ('90JZJrzlYX8wBpsCigK/0OYfPelYPYKE0AsWzyc0Kco=', '89287967-c45c-4f7a-862c-68560b45486d', 'postinstall', false, true),
  ('H2wNHQRIUrxtgTxPWlEPu6KDJDUoQRWvVFvw2dqwPeM=', '42a49084-171c-4801-a343-aaf91dd71d12', 'postinstall', false, true),
  ('Y6W8BIpGVPAt2cV9CBxgiAEuE+lIhtSnshuoknduWVM=', '1358f02b-0b67-40b4-b6fa-a6e2fd4afa63', 'postinstall', false, true),
  ('chNnfThlet39BP1DRo8FJmxCqdFAtvCcS3NN70Gxj84=', 'c9c275ba-d289-4537-9cdc-399cebdeaac1', 'postinstall', false, true);

-- Create extra files (QEMU scripts) for each package
INSERT INTO package_file (package_id, name, size, hash, hash256)
VALUES
  ('89287967-c45c-4f7a-862c-68560b45486d', 'oem-qemu.gz', '2078',
   'q1r8WHsEEKpJqpovT3ZbGbS5cxY=', '868966c4c50b96e4972ab86020dfd291fd3ca50700e212c4623d80b27c21f207'),
  ('42a49084-171c-4801-a343-aaf91dd71d12', 'oem-qemu.gz', '2059',
   'CHn6eoDzXLQ+llAbKiPRHTDi2wQ=', '79f3a34b4c869000294181d6f2430489c25d4abdee754a8430d8c7082b818d64'),
  ('1358f02b-0b67-40b4-b6fa-a6e2fd4afa63', 'oem-qemu.gz', '2040',
   'onBKiI5m4ltxxeBItUUdlSe7DsM=', 'df35f01071584d333218856920bf3d6dc66c69f32ab406c3b2addfcfaaf6371a'),
  ('c9c275ba-d289-4537-9cdc-399cebdeaac1', 'oem-qemu.gz', '2051',
   '0lKr62bjshV2M/OSlB2a9XOj+D8=', '612057711e2d6fd52799e92054f9f05e85bcc2f3a266e624f4faeaf099632006');

-- Point stable channel to latest version  
UPDATE channel SET package_id = 'c9c275ba-d289-4537-9cdc-399cebdeaac1' WHERE id = 'e06064ad-4414-4904-9a6e-fd465593d1b2';
EOF
```

## Test Scenarios

### A. Flatcar QEMU Instance Testing

#### Download and Setup QEMU Instance

````bash
# Create test directory
mkdir -p ~/flatcar-test && cd ~/flatcar-test

Follow the guide from the official docs to download to have a proper starting point with a version that is lower then all the other floors.
https://www.flatcar.org/docs/latest/installing/vms/qemu/

example:
```sh
# Download OLD Flatcar version (4081.3.5 LTS) for testing multi-step updates
wget https://lts.release.flatcar-linux.net/amd64-usr/4081.3.5/flatcar_production_qemu.sh
wget https://lts.release.flatcar-linux.net/amd64-usr/4081.3.5/flatcar_production_qemu_image.img.bz2
...
chmod +x flatcar_production_qemu.sh
````

# Create Butane config to point to local Nebraska

cat > nebraska-test.yaml <<EOF
variant: flatcar
version: 1.0.0
storage:
files: - path: /etc/flatcar/update.conf
contents:
inline: |
SERVER=http://10.0.2.2:8000
GROUP=stable
EOF

# Convert Butane to Ignition using Docker

docker run --rm -i quay.io/coreos/butane:latest < nebraska-test.yaml > nebraska-test.json

# Run Flatcar VM with snapshot mode (won't persist changes)

./flatcar_production_qemu.sh -i nebraska-test.json -nographic -snapshot

````

SSH into the instance (from another terminal):
```bash
ssh -o StrictHostKeyChecking=no -p 2222 core@localhost
````

#### Test Case 1: Direct Update (No Floors)

1. **Setup**: Ensure no floors are configured in the stable channel
2. **Action**: In the Flatcar VM, trigger update check:
   ```bash
   sudo update_engine_client -check_for_update
   sudo update_engine_client -status
   ```
3. **Expected**: Receives direct update from 4081.3.5 → 4230.2.2
4. **Verify**: Watch the update progress and check version after reboot

#### Test Case 2: Multi-Step Update with Floor Packages

1. **Setup via UI**:

   - Navigate to Nebraska UI at http://localhost:8002
   - Go to Channels → Edit stable channel
   - Add floor packages:
     - 4230.2.0 with a made up reason: "Breaking change in systemd requires intermediate update"
   - Save

2. **Action**: In Flatcar VM, trigger update:

   ```bash
   sudo update_engine_client -check_for_update
   sudo update_engine_client -status
   ```

3. **Expected Path**:

   - First update: 4081.3.5 → 4230.2.0 (floor)
   - After reboot, next update: 4230.2.0 → 4230.2.2 (target)

4. **Verify**:
   ```bash
   cat /etc/os-release | grep VERSION
   # Should show progression through versions
   ```

#### Test Case 3: Multiple Floors

1. **Setup via UI**:

   - Edit stable channel
   - Add multiple floors:
     - 4230.2.0 - "First migration step"
     - 4230.2.1 - "Security fix required before final update"

2. **Expected Update Path**:

   ```
   4081.3.5 → 4230.2.0 → 4230.2.1 → 4230.2.2
   ```

3. **Verify**: Each update completes and reboots before next step

### B. UI Testing

#### Test Case 4: Floor Package Management

1. **Via Package Edit Dialog**:

   - Navigate to Packages
   - Edit package 4230.2.0
   - In "Floor Channels" dropdown, select "stable"
   - Add floor reason: "Critical kernel update"
   - Save
   - Verify floor appears in channel

2. **Via Channel Edit Dialog**:

   - Navigate to Channels
   - Edit stable channel
   - View list of floor packages
   - Click delete icon to remove a floor
   - Verify removal

3. **Edge Cases**:
   - Try to blacklist a floor package (should fail)
   - Remove all floor channels from a package (reason should clear)

### C. API Testing

#### Test Case 5: Omaha Protocol Verification

Test the update response directly:

```bash
# Test from client at version 4081.3.5
curl -X POST http://localhost:8000/v1/update \
  -H "Content-Type: text/xml" \
  -d '<?xml version="1.0" encoding="UTF-8"?>
<request protocol="3.0" version="update_engine-0.4.2" ismachine="1">
  <os version="Chateau" platform="CoreOS" sp="4081.3.5_x86_64"></os>
  <app appid="e96281a6-d1af-4bde-9a0a-97b76e56dc57"
       version="4081.3.5"
       track="stable"
       board="amd64-usr">
    <updatecheck></updatecheck>
  </app>
</request>'
```

**Expected**: Response contains floor version (4230.2.0) not target (4230.2.2)

### D. Syncer Testing

#### Test Case 6: Syncer with Floors

```bash
# Set floors in UI first, then run syncer
docker run --rm \
  --network host \
  -e NEBRASKA_URL=http://localhost:8000 \
  -e SOURCE_URL=https://public.update.flatcar-linux.net \
  -e ARCH=amd64 \
  -e APP_ID=e96281a6-d1af-4bde-9a0a-97b76e56dc57 \
  flatcar/nebraska-syncer:latest
```

**Verify**: Floor settings preserved after sync

## Validation Checklist

- [ ] QEMU instance successfully updates through floor versions
- [ ] Updates happen in correct order (ascending version)
- [ ] Max 5 floors returned per update request
- [ ] UI correctly manages floor packages bidirectionally
- [ ] Floor reasons displayed and preserved
- [ ] Cannot blacklist floor packages
- [ ] Architecture filtering works correctly
- [ ] Syncer respects floor configurations

## Monitoring

```bash
# Watch Nebraska logs
tail -f nebraska.log | grep -E "floor|multi-step"

# Check floor configurations in database
PGPASSWORD=nebraska psql -h 127.0.0.1 -p 8001 -U postgres nebraska_tests \
  -c "SELECT c.name, p.version, cpf.floor_reason
      FROM channel_package_floors cpf
      JOIN channel c ON c.id = cpf.channel_id
      JOIN package p ON p.id = cpf.package_id
      ORDER BY p.version;"

# Monitor Flatcar update status
sudo journalctl -u update-engine -f
```
