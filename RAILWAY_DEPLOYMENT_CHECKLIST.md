# Railway Deployment Checklist - Image Persistence Fix

## Pre-Deployment

- [x] Added VOLUME declaration to Dockerfile
- [x] Updated docker-compose.yml to use named volume
- [x] Added storage verification function
- [x] Enhanced logging for debugging
- [x] Code compiles without errors

## Deployment Steps

### 1. Commit and Push Changes
```bash
git add backend/Dockerfile compose.yml backend/internal/storage/config.go backend/internal/server/server.go
git commit -m "Fix image persistence with Railway volume mount"
git push
```

### 2. Verify Railway Volume Configuration
In Railway dashboard:
- [ ] Navigate to your backend service
- [ ] Go to "Settings" → "Volumes"
- [ ] Verify volume exists with:
  - Name: `uploads`
  - Mount Path: `/app/uploads`
- [ ] If volume doesn't exist, create it with these settings

### 3. Monitor Deployment Logs
Watch for these success indicators:
```
✅ Storage config: UploadPath=/app/uploads, BaseURL=https://...
✅ Upload directories created successfully
✅ Storage path is writable - volume mount verified
```

Watch for these error indicators:
```
❌ CRITICAL: Failed to create upload directories
❌ CRITICAL: Storage path not writable
⚠️ This usually means the Railway volume is not properly mounted!
```

### 4. Test Image Upload (After Deployment)
- [ ] Upload a user avatar
- [ ] Verify image displays correctly
- [ ] Copy the image URL
- [ ] Trigger a redeploy (push empty commit or manual redeploy)
- [ ] After redeploy completes, verify the image still loads from the same URL

### 5. Verify Volume Persistence
```bash
# Option 1: Use Railway CLI
railway run bash
ls -la /app/uploads/
ls -la /app/uploads/avatars/
ls -la /app/uploads/poi-images/

# Option 2: Use Railway dashboard shell
# Click "Shell" in the service menu
ls -la /app/uploads/
```

## Troubleshooting

### Issue: "Storage path not writable" error

**Possible causes:**
1. Volume not created in Railway
2. Volume mount path mismatch
3. Permission issues

**Solutions:**
1. Check Railway dashboard → Volumes
2. Verify mount path is exactly `/app/uploads`
3. Delete and recreate the volume if needed
4. Check Railway service logs for mount errors

### Issue: Images still disappear after redeploy

**Possible causes:**
1. Volume not properly mounted
2. Files being written to wrong path
3. Multiple services writing to same volume

**Solutions:**
1. Check logs for storage path (should be `/app/uploads` on Railway)
2. Verify `RAILWAY_ENVIRONMENT` env var is set
3. SSH into container and check if `/app/uploads` is a mount point:
   ```bash
   mount | grep uploads
   df -h | grep uploads
   ```

### Issue: Volume fills up

**Solutions:**
1. Check volume size in Railway dashboard
2. Implement image cleanup for old/unused files
3. Consider upgrading volume size
4. Consider migrating to S3/cloud storage

## Post-Deployment Verification

- [ ] Images persist after redeploy
- [ ] No storage errors in logs
- [ ] Avatar uploads work correctly
- [ ] POI image uploads work correctly
- [ ] Image URLs are accessible
- [ ] Volume usage is reasonable

## Local Development Testing

Test the named volume locally:
```bash
# Start services
docker compose up -d

# Upload an image through the app

# Stop and remove containers (but keep volumes)
docker compose down

# Start again
docker compose up -d

# Verify image still exists
```

## Rollback Plan

If issues occur:
```bash
# Revert the changes
git revert HEAD
git push

# Or manually revert specific files
git checkout HEAD~1 -- backend/Dockerfile compose.yml backend/internal/storage/config.go backend/internal/server/server.go
git commit -m "Rollback image persistence changes"
git push
```

## Success Criteria

✅ All checks passed when:
- Deployment logs show storage verification success
- Images can be uploaded
- Images persist after redeploy
- No storage-related errors in logs
- Volume is properly mounted and writable

## Notes

- Railway volumes are persistent across deployments
- Volume data is backed up by Railway
- Volume size can be monitored in Railway dashboard
- Consider implementing storage monitoring and cleanup
