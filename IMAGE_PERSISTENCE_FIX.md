# Image Persistence Fix for Railway Deployments

## Problem
Images (user avatars and POI images) were disappearing after Railway redeploys because files were being written to the container's ephemeral filesystem instead of the persistent volume.

## Root Cause
The Dockerfile didn't declare `/app/uploads` as a VOLUME, which meant:
1. Files were written to the container's filesystem layer
2. Railway's volume mount might not take precedence over the container's directory
3. On redeploy, the container was rebuilt and all files in the container filesystem were lost

## Solution Applied

### 1. Dockerfile Changes (`backend/Dockerfile`)
Added explicit VOLUME declaration:
```dockerfile
# Declare volume mount point for persistent storage
# This ensures Railway's volume mount takes precedence
VOLUME ["/app/uploads"]
```

This tells Docker that `/app/uploads` is a mount point, ensuring Railway's persistent volume takes precedence.

### 2. Docker Compose Changes (`compose.yml`)
Changed from bind mount to named volume for local development:
```yaml
volumes:
  - uploads_data:/app/uploads  # Named volume (persistent)

volumes:
  uploads_data:  # Declared at bottom
```

This ensures local development also uses persistent storage.

### 3. Storage Verification (`backend/internal/storage/config.go`)
Added `VerifyStorageWritable()` function that:
- Tests write permissions on startup
- Helps diagnose volume mount issues
- Provides clear error messages if volume isn't mounted

### 4. Enhanced Logging (`backend/internal/server/server.go`)
Added detailed logging during startup:
- Storage configuration (path and base URL)
- Directory creation status
- Write permission verification
- Clear warnings if volume mount fails

## Verification Steps for Railway

### 1. Check Volume Configuration
In Railway dashboard:
- Go to your service settings
- Verify the volume is listed under "Volumes"
- Confirm mount path is `/app/uploads`
- Note the volume ID

### 2. Check Deployment Logs
After deploying, look for these log messages:
```
✅ Storage config: UploadPath=/app/uploads, BaseURL=https://...
✅ Upload directories created successfully
✅ Storage path is writable - volume mount verified
```

If you see errors like:
```
❌ CRITICAL: Storage path not writable
⚠️ This usually means the Railway volume is not properly mounted!
```
Then the volume isn't working.

### 3. Test Image Upload
1. Upload a user avatar or POI image
2. Verify the image displays correctly
3. Note the image URL (should be like: `https://your-domain.railway.app/uploads/avatars/...`)
4. Trigger a redeploy (push a commit or manual redeploy)
5. After redeploy, verify the image still loads

### 4. Inspect Volume Contents (if needed)
You can SSH into the Railway container to inspect:
```bash
# In Railway dashboard, use the "Shell" feature
ls -la /app/uploads/
ls -la /app/uploads/avatars/
ls -la /app/uploads/poi-images/
```

## How Railway Volumes Work

Railway persistent volumes:
- Survive container restarts and redeploys
- Are mounted at the specified path
- Persist data across deployments
- Are backed up by Railway

The VOLUME declaration in the Dockerfile ensures:
- Docker knows this is a mount point
- The directory is treated specially during container creation
- Railway's volume mount takes precedence over any directory in the image

## Local Development

The named volume `uploads_data` in docker-compose ensures:
- Files persist between `docker compose down` and `docker compose up`
- Files are NOT lost when containers are recreated
- Files are only lost with `docker compose down -v` (removes volumes)

To inspect local volume:
```bash
docker volume inspect breakoutglobe_uploads_data
docker run --rm -v breakoutglobe_uploads_data:/data alpine ls -la /data
```

## Rollback Plan

If issues occur, you can temporarily revert by:
1. Removing the VOLUME declaration from Dockerfile
2. Reverting compose.yml changes
3. Using environment variable to set a different upload path

## Next Steps

After deploying:
1. Monitor the startup logs for storage verification messages
2. Test image upload and persistence
3. Trigger a test redeploy to verify images survive
4. Consider adding a health check endpoint that verifies storage is writable

## Additional Considerations

### Future Improvements
- Add periodic health checks for storage writability
- Implement storage usage monitoring
- Add automatic cleanup of old/unused images
- Consider migrating to S3/cloud storage for better scalability

### Monitoring
Watch for these issues:
- Volume running out of space
- Permission errors in logs
- Images not loading after redeploy
- Slow file access (volume performance)
