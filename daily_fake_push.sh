#!/bin/bash

echo "Script ran at $(date)" >> /home/tejsva/cron_debug.txt


# Go to your project directory
cd /home/tejsva/Desktop/coding/docmake

# Update a dummy file so that Git has something to commit
echo "Update: $(date '+%Y-%m-%d %H:%M:%S')" >> daily_log.txt

# Stage changes
git add .

# Commit with timestamp
git commit -m "Daily auto-commit: $(date '+%Y-%m-%d')"

# Push to GitHub
git push origin main
