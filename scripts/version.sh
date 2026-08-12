#!/bin/sh
# Version rule: <major>.<minor>.<patch> (e.g. 0.0.09), derived from the git
# commit count so every commit bumps the patch by 1 automatically. Patch is
# zero-padded to 2 digits and carries at 99: 0.0.99 -> 0.1.00 -> ... -> 0.99.99
# -> 1.0.00.
set -e
n=$(git rev-list --count HEAD 2>/dev/null || echo 0)
major=$((n / 10000))
minor=$(((n % 10000) / 100))
patch=$((n % 100))
printf "%d.%d.%02d" "$major" "$minor" "$patch"
