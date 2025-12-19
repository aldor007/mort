# S3 GLACIER Object Restore

Automatic restoration of archived S3 objects from GLACIER and DEEP_ARCHIVE storage classes.

## Overview

When S3 objects are moved to GLACIER or DEEP_ARCHIVE storage classes (typically via lifecycle policies), they cannot be accessed directly. Attempting to retrieve them results in an `InvalidObjectState` error.

Mort now automatically handles GLACIER objects by:
1. Detecting `InvalidObjectState` errors
2. Initiating S3 restore requests via the AWS API
3. Tracking restore status to prevent duplicate requests
4. Returning `503 Service Unavailable` with `Retry-After` header
5. Serving the object once restore completes

## Configuration

GLACIER restore is **disabled by default** and must be explicitly enabled per bucket.

### Basic Configuration

```yaml
buckets:
  media:
    glacier:
      enabled: true              # REQUIRED: Must be set to true to enable
      restoreTier: "Standard"    # Optional: Expedited, Standard, or Bulk
      restoreDays: 7             # Optional: How long restored object stays available
      retryAfterSeconds: 14400   # Optional: Auto-calculated based on tier
    storages:
      basic:
        kind: "s3"
        accessKey: "${S3_ACCESS_KEY}"
        secretAccessKey: "${S3_SECRET_KEY}"
        region: "us-east-1"
        bucket: "my-bucket"
```

### Configuration Options

#### `enabled` (boolean, default: `false`)
**REQUIRED:** Must be set to `true` to enable automatic GLACIER restore.

If disabled or not configured:
- GLACIER errors return `503 Service Unavailable` without initiating restore
- Objects remain inaccessible until manually restored via AWS Console/CLI

#### `restoreTier` (string, default: `"Standard"`)
Controls restore speed and cost:

| Tier | Restore Time | Cost | Use Case |
|------|-------------|------|----------|
| `Expedited` | 1-5 minutes | $$$ | Production serving, user-facing |
| `Standard` | 3-5 hours | $$ | Balanced cost/performance |
| `Bulk` | 5-12 hours | $ | Batch operations, non-urgent |

**Note:** Expedited tier availability depends on AWS capacity. May fail during high demand.

#### `restoreDays` (integer, default: `3`)
Number of days the restored copy remains accessible (1-30).

After expiration, object returns to GLACIER and must be restored again.

**Recommendations:**
- **7 days** - Good balance for frequently accessed images
- **30 days** - For consistently accessed content (higher cost)
- **1 day** - Minimal cost, for one-time access

#### `retryAfterSeconds` (integer, auto-calculated)
Value for `Retry-After` HTTP header telling clients when to retry.

**Auto-calculated based on tier:**
- Expedited: 300 seconds (5 minutes)
- Standard: 14400 seconds (4 hours)
- Bulk: 43200 seconds (12 hours)

Override only if you need custom retry intervals.

---

## How It Works

### Request Flow

```
Client Request
    ↓
Mort receives request for GLACIER object
    ↓
Storage layer: item.Open() fails with InvalidObjectState
    ↓
GLACIER detection: Error pattern matches
    ↓
Check restore cache: Already requested?
    ├─ Yes → Return 503 with cached status
    └─ No  → Initiate restore (async)
         ├─ Call S3 RestoreObject API
         ├─ Mark in cache (prevent duplicates)
         └─ Log & increment metrics
    ↓
Return 503 Service Unavailable
    ↓
Client sees Retry-After header, waits
    ↓
Client retries after specified time
    ↓
Object available (restore completed)
    ↓
Normal image serving resumes
```

### HTTP Response

When a GLACIER object is requested:

```http
HTTP/1.1 503 Service Unavailable
Retry-After: 14400
X-Mort-Glacier-Status: restoring
X-Mort-Glacier-Tier: Standard
Content-Type: application/json

{"message": "object in GLACIER storage class, restore in progress"}
```

**Headers:**
- `Retry-After`: Seconds to wait before retrying (based on restore tier)
- `X-Mort-Glacier-Status`: Always "restoring"
- `X-Mort-Glacier-Tier`: The tier used (Expedited/Standard/Bulk)

---

## Restore Cache

Mort tracks restore requests to avoid duplicate S3 API calls.

### Cache Backends

**Memory Cache** (default):
- Per-instance tracking
- Lost on restart
- Suitable for single-instance deployments

**Redis Cache** (recommended for production):
- Shared across all mort instances
- Persists across restarts
- Prevents duplicate restores in multi-instance setup

### Configuration

Use your existing cache configuration:

```yaml
server:
  cache:
    type: "redis"              # or "memory"
    address:
      - "localhost:6379"
```

The GLACIER restore cache automatically uses the same backend.

**Redis Key Format:** `mort:glacier:restore:<object-key>`
**TTL:** `retryAfterSeconds + 24 hours` (to track completion)

---

## Cost Considerations

### Restore Costs

S3 GLACIER restore pricing (varies by region):

| Tier | Per GB | Per 1000 Requests | Typical Image Cost |
|------|--------|-------------------|-------------------|
| Expedited | $0.03 | $10.00 | ~$0.001 per 1MB image |
| Standard | $0.01 | $0.05 | ~$0.0001 per 1MB image |
| Bulk | $0.0025 | $0.025 | ~$0.00003 per 1MB image |

### Storage Costs

Once restored, object exists in BOTH GLACIER and Standard:
- **GLACIER copy:** $0.004/GB/month (remains)
- **Standard copy:** $0.023/GB/month (temporary, expires after `restoreDays`)

**Example:** 100GB of images restored for 7 days:
- GLACIER storage: $0.40/month (ongoing)
- Temporary Standard storage: ~$0.53 for 7 days
- Restore cost (Standard tier): ~$1.00 one-time

### Cost Optimization Tips

1. **Use Bulk tier** for non-urgent restores (5x cheaper than Standard)
2. **Set appropriate restoreDays** - don't use 30 days if 7 is sufficient
3. **Monitor `glacier_restore_initiated` metric** - high frequency means lifecycle policy too aggressive
4. **Consider Intelligent-Tiering** instead of GLACIER for frequently accessed data

---

## Monitoring

### Prometheus Metrics

```go
glacier_error_detected         // Count of GLACIER errors encountered
glacier_restore_initiated      // Count of restore requests sent to S3
```

### Log Events

**Restore Initiated:**
```json
{
  "level": "info",
  "msg": "Initiating GLACIER restore",
  "obj.Key": "/path/to/image.jpg",
  "tier": "Standard",
  "days": 7
}
```

**Restore Already in Progress:**
```json
{
  "level": "info",
  "msg": "GLACIER restore already in progress (cached)",
  "obj.Key": "/path/to/image.jpg",
  "requestedAt": "2025-12-18T10:00:00Z",
  "expiresAt": "2025-12-18T14:00:00Z"
}
```

**Restore Failed:**
```json
{
  "level": "error",
  "msg": "GLACIER restore failed",
  "obj.Key": "/path/to/image.jpg",
  "error": "AccessDenied"
}
```

---

## IAM Permissions

The S3 user/role must have `s3:RestoreObject` permission:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:RestoreObject"
      ],
      "Resource": "arn:aws:s3:::my-bucket/*"
    }
  ]
}
```

---

## Testing

### Enable for Test Bucket

```yaml
buckets:
  test:
    glacier:
      enabled: true
      restoreTier: "Expedited"  # Faster for testing
      restoreDays: 1            # Minimal cost
```

### Simulate GLACIER Request

1. Move test object to GLACIER via AWS CLI:
   ```bash
   aws s3api copy-object \
     --bucket my-bucket \
     --key test/image.jpg \
     --copy-source my-bucket/test/image.jpg \
     --storage-class GLACIER \
     --metadata-directive COPY
   ```

2. Request the object via Mort:
   ```bash
   curl -i http://mort-server/test/image.jpg
   ```

3. Expect `503` response with `Retry-After` header

4. Check logs for "Initiating GLACIER restore"

5. Wait for `Retry-After` duration and retry

---

## Troubleshooting

### Object Still Returns 503 After Retry-After

**Cause:** Restore may take longer than estimated.

**Solution:**
- Check S3 console for actual restore status
- Increase `retryAfterSeconds` in config
- For Expedited tier, AWS may throttle during high demand

### Restore Metric Not Incrementing

**Cause:** GLACIER detection not working.

**Check:**
- Is `glacier.enabled: true` in config?
- Is error actually `InvalidObjectState`?
- Check logs for "GLACIER error detected"

### Too Many Restore Requests

**Cause:** Restore cache not working or TTL too short.

**Solution:**
- Use Redis cache (shared across instances)
- Verify cache is accessible: `redis-cli ping`
- Check `retryAfterSeconds` is appropriate for tier

### High GLACIER Restore Costs

**Cause:** Too many objects in GLACIER being requested.

**Solutions:**
- Review S3 lifecycle policies - may be too aggressive
- Use Bulk tier instead of Expedited (5x cost savings)
- Reduce `restoreDays` to minimum needed
- Consider moving frequently accessed objects to Standard storage

---

## Migration Guide

### Existing Deployments

GLACIER restore is **disabled by default** - no changes needed for existing deployments.

To enable:

1. Update config.yml:
   ```yaml
   buckets:
     your-bucket:
       glacier:
         enabled: true
         restoreTier: "Standard"
         restoreDays: 7
   ```

2. Deploy updated configuration

3. Monitor `glacier_error_detected` and `glacier_restore_initiated` metrics

4. Adjust `restoreTier` and `restoreDays` based on cost/performance needs

### New Deployments

For buckets with GLACIER lifecycle policies:

1. Enable GLACIER restore in config
2. Ensure IAM has `s3:RestoreObject` permission
3. Use Redis cache for multi-instance setups
4. Set appropriate `restoreTier` based on SLA requirements

---

## FAQ

**Q: Does this work with S3-compatible storage (MinIO, DigitalOcean Spaces)?**
A: Only if the storage supports GLACIER storage classes and RestoreObject API. Most S3-compatible storage doesn't support GLACIER.

**Q: What happens if restore fails?**
A: Error is logged. Client still receives 503. Restore will be reattempted on next request (if not in cache).

**Q: Can I manually restore objects instead?**
A: Yes. Set `enabled: false` or omit `glacier` config entirely. Restore via AWS Console/CLI, then objects are accessible.

**Q: Does this affect performance?**
A: No impact on non-GLACIER objects. GLACIER error detection adds ~1ms overhead only when error occurs.

**Q: How do I know if restore completed?**
A: Client receives 503 with `Retry-After`. After waiting, retry the request. If restore complete, mort serves the image normally (200 OK).

**Q: What if I have both transformed and original images in GLACIER?**
A: Both are handled independently. Original image restores first, then transformation can proceed.

---

## Examples

### Minimal Configuration

```yaml
buckets:
  photos:
    glacier:
      enabled: true  # Use all defaults (Standard, 3 days, 4-hour retry)
    storages:
      basic:
        kind: "s3"
        region: "us-east-1"
```

### Production Configuration

```yaml
buckets:
  production-images:
    glacier:
      enabled: true
      restoreTier: "Standard"      # Balanced cost/performance
      restoreDays: 7               # One week availability
      retryAfterSeconds: 10800     # 3 hours (conservative)
    storages:
      basic:
        kind: "s3"
        region: "us-east-1"
        accessKey: "${S3_ACCESS_KEY}"
        secretAccessKey: "${S3_SECRET_KEY}"
        bucket: "production-bucket"
      transform:
        kind: "s3"
        region: "us-east-1"
        accessKey: "${S3_ACCESS_KEY}"
        secretAccessKey: "${S3_SECRET_KEY}"
        bucket: "production-transforms"

server:
  cache:
    type: "redis"
    address:
      - "redis-master:6379"
      - "redis-replica:6379"
```

### High-Performance Configuration

```yaml
buckets:
  critical:
    glacier:
      enabled: true
      restoreTier: "Expedited"     # 1-5 minute restore (expensive!)
      restoreDays: 1               # Minimal storage cost
      retryAfterSeconds: 300       # 5 minutes
    storages:
      basic:
        kind: "s3"
        region: "us-east-1"
```

---

## Implementation Details

### Architecture

- **Stow Library:** Restorable interface implemented in S3 backend
- **Mort Integration:** GLACIER error detection and restore orchestration
- **Cache Layer:** Shared restore status tracking

### Error Detection

Pattern matching on error messages:
- Checks for "InvalidObjectState" string
- Validates storage class (GLACIER, DEEP_ARCHIVE)
- Only applies to S3 storage backend

### Restore Process

1. **Detection:** `item.Open()` fails with InvalidObjectState
2. **Cache Check:** Query restore cache by object key
3. **Initiate:** Call stow `Restorable.Restore(ctx, days, tier)` async
4. **Track:** Mark restore request in cache with TTL
5. **Respond:** Return 503 with headers

### Thread Safety

- Restore initiation happens in background goroutine (non-blocking)
- Cache operations are thread-safe (sync.RWMutex for memory, Redis for distributed)
- Multiple requests for same object handled via cache (only one restore request sent)

---

## Changelog

### v0.35.0 (2025-12-18)

**New Features:**
- Added S3 GLACIER/DEEP_ARCHIVE automatic restore
- Configurable restore tier and duration per bucket
- Restore status tracking via Redis/Memory cache
- 503 responses with Retry-After headers

**Dependencies:**
- Updated `github.com/aldor007/stow` to v1.5.0 (GLACIER support)

**Configuration:**
- Added `glacier` section to bucket configuration
- Disabled by default (opt-in per bucket)

---

## Support

For issues or questions:
- GitHub Issues: https://github.com/aldor007/mort/issues
- Check logs for "GLACIER" keyword
- Monitor Prometheus metrics: `glacier_*`
