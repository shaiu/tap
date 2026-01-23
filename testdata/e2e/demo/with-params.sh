#!/bin/bash
# ---
# name: with-params
# description: Script with parameters for integration testing
# category: demo
# parameters:
#   - name: name
#     type: string
#     required: true
#     description: Name to greet
#   - name: count
#     type: int
#     default: 1
#     description: Number of times to greet
#   - name: loud
#     type: bool
#     default: false
#     description: Use uppercase
#   - name: env
#     type: string
#     choices: [dev, staging, prod]
#     default: dev
#     description: Environment
# ---

msg="Hello, ${TAP_PARAM_NAME}! (env: ${TAP_PARAM_ENV})"

i=1
while [ "$i" -le "${TAP_PARAM_COUNT:-1}" ]; do
    if [ "$TAP_PARAM_LOUD" = "true" ]; then
        echo "$msg" | tr '[:lower:]' '[:upper:]'
    else
        echo "$msg"
    fi
    i=$((i + 1))
done

exit 0
