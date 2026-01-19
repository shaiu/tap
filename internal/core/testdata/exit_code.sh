#!/bin/bash
# ---
# name: exit-code-test
# description: Test script that exits with a specific code
# category: test
# ---
exit ${TAP_PARAM_CODE:-0}
