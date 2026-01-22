#!/bin/bash
# ---
# name: show-env
# description: Display TAP environment variables
# category: utils
# ---
echo "=== TAP Environment Variables ==="
env | grep "^TAP_" || echo "No TAP_ variables set"
echo ""
echo "Script: $TAP_SCRIPT_NAME"
echo "Path: $TAP_SCRIPT_PATH"
exit 0
