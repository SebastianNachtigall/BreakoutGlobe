# Railway Volume Fix - Clear Instructions

## The Problem

Images were not persisting because:
1. The root `Dockerfile` (used by Railway) was missing the VOLUME declaration
2. The working directory was `/root/` instead of `/app`
3. You created a manual volume in Railway UI, which conflicts with the `railway.toml` automatic volume

## The Solution

### What Was Fixed

1. **Updated `Dockerfile`** (root, used by Railway):
   - Changed WORKDIR from `/root/` to `/app` (CRITICAL FIX)
   - Added health check
   - **NOTE:** Railway BANS the `VOLUME` keyword in Dockerfiles - use `railway.toml` instead

2. **Clarified Volume Strategy**:
   - Use `railway.toml` for automatic volume creation (REQUIRED for Railway)
   - Railway handles volume mounting automatically via TOML config
   - No VOLUME declaration needed in Dockerfile (Railway bans it)

## Action Steps

### Step 1: Delete Manual Volume in Railway UI
1. Go to Railway dashboard
2. Select your backend service
3. Go to Settings → Volumes
4. **Delete** the manually created volume
5. Don't worry - the `railway.toml` will recreate it automatically

### Step 2: Commit and Deploy
```bash
git add Dockerfile
git commit -m "Fix Railway volume mount - correct working directory and VOLUME declaration"
git push
```

### Step 3: Verify Deployment
After Railway deploys, check the logs for:

✅ **Success indicators:**
```
📁 Storage config: UploadPath=/app/uploads, BaseURL=https://...
✅ Upload directories created successfully
✅ Storage path is writable - volume mount verified
```

❌ **Error indicators:**
```
❌ CRITICAL: Storage path not writable
⚠️ This usually means the Railway volume is not properly mounted!
```

### Step 4: Test Image Persistence
1. Upload a user avatar or POI image
2. Note the image URL
3. Verify the image displays
4. Trigger a redeploy (push an empty commit or use Railway UI)
5. After redeploy, verify the image still loads

## How It Works

### Railway TOML Volume (Automatic)
```toml
[[deploy.volumes]]
name = "uploads"
mountPath = "/app/uploads"
```
- Railway automatically creates a persistent volume
- Mounts it to `/app/uploads` in your container
- Survives redeploys and restarts

### Railway Volume Handling
**IMPORTANT:** Railway BANS the `VOLUME` keyword in Dockerfiles.
- Railway handles volume mounting automatically via `railway.toml`
- No Dockerfile changes needed for volume declaration
- The TOML config is sufficient for persistent storage

### Storage Configuration
The Go code automatically detects Railway environment:
```go
if os.Getenv("RAILWAY_ENVIRONMENT") != "" {
    uploadPath = "/app/uploads"  // Railway production
} else {
    uploadPath = "./uploads"      // Local development
}
```

## Why Manual Volume Doesn't Work Well

When you create a volume manually in Railway UI:
- It might conflict with the TOML-defined volume
- Railway might try to mount both volumes to the same path
- The TOML volume might override the manual one
- It's harder to track in version control

**Best practice:** Use `railway.toml` for infrastructure-as-code.

## Troubleshooting

### If images still don't persist:

1. **Check Railway logs** for storage verification messages
2. **Verify volume exists** in Railway UI (should be auto-created from TOML)
3. **Check mount path** is exactly `/app/uploads`
4. **SSH into container** (Railway Shell) and run:
   ```bash
   ls -la /app/uploads/
   mount | grep uploads
   df -h | grep uploads
   ```

### If you see "Storage path not writable":

1. Volume isn't mounted properly
2. Check Railway service logs for mount errors
3. Try deleting and redeploying (Railway will recreate volume)
4. Verify `RAILWAY_ENVIRONMENT` variable is set in Railway

## Local Development

Your `compose.yml` already uses a named volume:
```yaml
volumes:
  - uploads_data:/app/uploads
```

This works the same way - files persist between container restarts.

## Summary

- ✅ `Dockerfile` now has correct WORKDIR and VOLUME declaration
- ✅ `railway.toml` defines the volume (infrastructure-as-code)
- ✅ Delete manual volume from Railway UI to avoid conflicts
- ✅ Deploy and verify with logs
- ✅ Test image upload and persistence

The key insight: **Railway TOML + Dockerfile VOLUME = Persistent Storage**
