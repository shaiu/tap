#!/bin/bash
# ---
# name: stderr-test
# description: Test script that writes to stderr
# category: test
# ---
echo "stdout message"
echo "stderr message" >&2
exit 0
