# Building with GitHub Actions (Zero Local Build Time!)

## 🎉 Good News!

GitHub Actions is **already configured** in this repository! You can trigger builds in the cloud and download the results - **no local build time required**.

## Current GitHub Actions Setup

### 1. CI Workflow (ci.yml)
- **Triggers**: Automatically on push to `master` or pull requests
- **Runs on**: `ubuntu-24.04-arm` (native ARM runners - fast!)
- **Builds**: Single .deb package via `make ddpkg`
- **Time**: ~15-20 minutes
- **Artifacts**: Downloadable .deb package

### 2. Release Workflow (release.yml)
- **Triggers**: When you push a version tag (e.g., `v1.6.0`)
- **Runs on**: `ubuntu-24.04-arm` (native ARM runners)
- **Builds**: Full Raspberry Pi image + .deb package
- **Time**: ~45-60 minutes
- **Artifacts**: Both .img and .deb files as GitHub Release

## Option 1: Trigger Build Manually (EASIEST)

### Method A: Via GitHub Website
1. Go to https://github.com/montge/stratux/actions
2. Click on "CI" workflow
3. Click "Run workflow" button
4. Select branch (master)
5. Click green "Run workflow"
6. Wait 15-20 minutes
7. Download .deb from "Artifacts" section

### Method B: Via `gh` CLI (You Have This!)
```bash
# Trigger CI build from command line
gh workflow run ci.yml

# Watch the progress
gh run watch

# List recent runs
gh run list --workflow=ci.yml

# Download the artifact when done
gh run download <run-id>
```

## Option 2: Push to GitHub (Automatic Build)

Every time you push to master, GitHub automatically builds:

```bash
# Make sure you're on master
git checkout master

# Commit your changes (if any)
git add .
git commit -m "Your changes"

# Push to trigger automatic build
git push origin master

# Watch the build
gh run watch
```

## Option 3: Create a Release (Full Image Build)

To build full Raspberry Pi images:

```bash
# Create and push a tag
git tag v1.6.0-test
git push origin v1.6.0-test

# This triggers release.yml which builds:
# - Full .img file
# - .deb package
# Both uploaded as a GitHub Release (draft)

# Watch progress
gh run watch

# When done, view the release
gh release view v1.6.0-test
```

## Comparing Your Options

| Method | Time (You) | Time (GitHub) | Cost | Output |
|--------|-----------|---------------|------|--------|
| **GitHub Actions** | 2 min | 15-20 min | FREE | .deb package |
| **Local Cross-Compile** | 30-60 min | 0 min | FREE | .deb + .img |
| **Native on Crewdog** | 5-15 min | 0 min | FREE | .deb package |

## Recommended: GitHub Actions for Most Builds!

**Why use GitHub Actions:**
- ✅ **Zero local build time** - do other work while it builds
- ✅ **Native ARM runners** - fast, not emulated
- ✅ **Free for public repos** (2000 min/month for private)
- ✅ **Reproducible** - same environment every time
- ✅ **Automatic** - builds on every push
- ✅ **Multiple builds** - can run US and EU in parallel

**When to use local builds:**
- Quick testing on crewdog (5-15 min native build)
- No internet connection
- Want full control over build environment

## Creating US and EU Builds with GitHub Actions

Currently, the workflow builds one default config. Let me create a workflow that builds both:

### Option A: Manual Trigger for Both Regions

I can create a new workflow that lets you choose:

```bash
# Trigger build with region selection
gh workflow run build-regions.yml -f region=US
gh workflow run build-regions.yml -f region=EU

# Or both at once
gh workflow run build-regions.yml -f region=both
```

### Option B: Automatic on Push

Modify ci.yml to build both US and EU automatically on every push.

## Quick Start: Using GitHub Actions NOW

### 1. Trigger a Test Build
```bash
# From your repo directory
cd /home/e/Development/stratux

# Trigger the CI build
gh workflow run ci.yml

# Watch it in real-time
gh run watch
```

### 2. When Build Completes (15-20 min)
```bash
# List recent runs to get the run ID
gh run list --workflow=ci.yml --limit 1

# Download the artifact (replace <run-id> with actual ID)
gh run download <run-id>

# This creates a directory with the .deb file
# Copy to crewdog
scp stratux-debian-package/stratux-*.deb pi@10.0.1.53:~/
```

### 3. Deploy to Crewdog
```bash
ssh pi@10.0.1.53
sudo dpkg -i stratux-*.deb
sudo systemctl restart stratux
```

## Monitoring Builds

### Watch Progress
```bash
# Watch the currently running build
gh run watch

# Or watch a specific run
gh run watch <run-id>
```

### Check Status
```bash
# List recent workflow runs
gh run list --workflow=ci.yml

# View details of a specific run
gh run view <run-id>
```

### View Logs
```bash
# View logs from a run
gh run view <run-id> --log

# View logs and follow (live)
gh run view <run-id> --log-failed
```

## Creating Multi-Region Workflow

Would you like me to create a new GitHub Actions workflow that:
- Builds both US and EU versions
- Can be manually triggered
- Uploads both as separate artifacts
- Takes ~20-25 minutes total (builds in parallel)

Let me know and I'll create it!

## Summary

**Recommended approach:**
1. **Use GitHub Actions for production builds** (trigger with `gh workflow run ci.yml`)
2. **Use native builds on crewdog for testing** (5-15 min for quick iterations)
3. **Download from GitHub Actions and deploy** (zero local build time!)

**Right now, you can:**
```bash
# Start a build on GitHub
gh workflow run ci.yml

# Do other work for 15 minutes...

# Check if done
gh run list --workflow=ci.yml

# Download when ready
gh run download <run-id>

# Deploy to crewdog
scp stratux-debian-package/stratux-*.deb pi@10.0.1.53:~/
```

**This is much easier than waiting 60 minutes locally!**
