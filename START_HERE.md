# 🚀 Stratux Build System - START HERE

## The Easiest Way: GitHub Actions (Recommended!)

**You already have GitHub Actions configured!** Let GitHub build for you:

### ⚡ Quick Start (2 minutes of your time)

```bash
# 1. Push the new workflow to GitHub
cd /home/e/Development/stratux
git add .github/workflows/build-regions.yml
git commit -m "Add US/EU region build workflow"
git push origin master

# 2. Trigger a build (pick one region or both)
gh workflow run build-regions.yml -f region=both

# 3. Watch it build (optional)
gh run watch

# 4. After 15-20 minutes, download results
gh run list --workflow=build-regions.yml --limit 1
# Note the run ID, then:
gh run download <run-id>

# 5. Deploy to crewdog
scp stratux-*-debian-package/stratux-*.deb pi@10.0.1.53:~/
ssh pi@10.0.1.53 "sudo dpkg -i stratux-*.deb && sudo systemctl restart stratux"
```

**Your time**: 2 minutes to trigger + 2 minutes to download/deploy
**GitHub's time**: 15-20 minutes (do other things while it builds!)

---

## Alternative Methods

### If You Need It NOW (Fastest)
Build directly on crewdog:
```bash
# Copy repo to crewdog and build there (5-15 min)
# See: DEPLOY_TO_CREWDOG.md
```

### If You Want Full Control
Build locally with Docker:
```bash
# Takes 30-60 minutes on your machine
# See: BUILD_INSTRUCTIONS.md
```

---

## Decision Tree

```
Do you need the build RIGHT NOW?
├─ YES → Build on crewdog (5-15 min)
│        See: DEPLOY_TO_CREWDOG.md
│
└─ NO  → Use GitHub Actions (15-20 min, zero local time)
         You're reading the right guide!
```

---

## Using GitHub Actions - Step by Step

### Step 1: Commit and Push the New Workflow
```bash
cd /home/e/Development/stratux

# Add the new workflow file
git add .github/workflows/build-regions.yml

# Commit
git commit -m "Add US/EU region build workflow"

# Push to trigger
git push origin master
```

### Step 2: Trigger a Build

**Option A: Build Both Regions (Recommended)**
```bash
gh workflow run build-regions.yml -f region=both
```

**Option B: Build Just US**
```bash
gh workflow run build-regions.yml -f region=US
```

**Option C: Build Just EU**
```bash
gh workflow run build-regions.yml -f region=EU
```

**Option D: Use the Web Interface**
1. Go to https://github.com/montge/stratux/actions
2. Click "Build US/EU Regions"
3. Click "Run workflow"
4. Select region
5. Click green "Run workflow"

### Step 3: Monitor Progress (Optional)

```bash
# Watch the build in real-time
gh run watch

# Or check status
gh run list --workflow=build-regions.yml
```

### Step 4: Download When Complete (~15-20 min later)

```bash
# List recent runs to get ID
gh run list --workflow=build-regions.yml --limit 1

# Download artifacts (replace <run-id>)
gh run download <run-id>

# This creates directories:
# - stratux-US-debian-package/stratux-US-*.deb
# - stratux-EU-debian-package/stratux-EU-*.deb
```

### Step 5: Deploy to Crewdog

```bash
# Copy to crewdog
scp stratux-US-debian-package/stratux-US-*.deb pi@10.0.1.53:~/

# Install
ssh pi@10.0.1.53 "sudo dpkg -i stratux-US-*.deb && sudo systemctl restart stratux"

# Verify
ssh pi@10.0.1.53 "sudo systemctl status stratux"
```

---

## What Gets Built

### When You Run `build-regions.yml`:

**If you choose "both":**
- ✅ US .deb package (UAT enabled)
- ✅ EU .deb package (OGN enabled)
- ⏱️ Time: ~20 minutes (parallel builds)
- 📦 Two separate artifacts to download

**If you choose "US" or "EU":**
- ✅ One region's .deb package
- ⏱️ Time: ~15 minutes
- 📦 One artifact to download

---

## Comparison of All Methods

| Method | Your Time | Total Time | Output | Best For |
|--------|-----------|------------|--------|----------|
| **GitHub Actions** | 2 min | 15-20 min | Both US & EU .deb | Production builds |
| **Native (crewdog)** | 5-15 min | 5-15 min | One .deb | Quick testing |
| **Local Cross-compile** | 30-60 min | 30-60 min | Both .deb + .img | Full control |

---

## FAQ

**Q: Do I need to wait during the GitHub build?**
A: No! Trigger it and do other work. Download when you're ready.

**Q: Does GitHub Actions cost money?**
A: Free for public repos. Private repos get 2000 minutes/month free.

**Q: Can I build both US and EU at once?**
A: Yes! Choose "both" when running the workflow. They build in parallel.

**Q: How do I know when the build is done?**
A: Run `gh run list --workflow=build-regions.yml` or check the website.

**Q: What if the build fails?**
A: Run `gh run view <run-id> --log` to see error logs.

**Q: Can I automate this on every push?**
A: Yes, but currently it's manual trigger only. Let me know if you want automatic.

---

## Next Steps

1. **Right now**: Push the new workflow and trigger a build
2. **While it builds**: Read other docs or work on something else
3. **After 15-20 min**: Download and deploy to crewdog
4. **For quick testing**: Use native builds on crewdog (see DEPLOY_TO_CREWDOG.md)

---

## All Documentation

- **[START_HERE.md](START_HERE.md)** ← You are here
- **[GITHUB_ACTIONS_BUILD.md](GITHUB_ACTIONS_BUILD.md)** - Detailed GitHub Actions guide
- **[DEPLOY_TO_CREWDOG.md](DEPLOY_TO_CREWDOG.md)** - Native build on your Pi
- **[WHICH_BUILD_METHOD.md](WHICH_BUILD_METHOD.md)** - Compare all methods
- **[BUILD_INSTRUCTIONS.md](BUILD_INSTRUCTIONS.md)** - Complete reference
- **[CLAUDE.md](CLAUDE.md)** - Architecture overview

---

## TL;DR

```bash
# Push new workflow
git add .github/workflows/build-regions.yml
git commit -m "Add region build workflow"
git push

# Build both US and EU (15-20 min)
gh workflow run build-regions.yml -f region=both

# Download when done
gh run download <run-id>

# Deploy
scp stratux-*-debian-package/*.deb pi@10.0.1.53:~/
ssh pi@10.0.1.53 "sudo dpkg -i stratux-*.deb && sudo systemctl restart stratux"
```

**Done! Let GitHub do the heavy lifting!** 🎉
