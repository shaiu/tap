#!/bin/bash
# Ralph Wiggum loop for tap
# Usage: ./loop.sh [plan|build]

set -e

MODE="${1:-build}"
MAX_ITERATIONS="${2:-20}"
ITERATION=0

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}═══════════════════════════════════════${NC}"
echo -e "${GREEN}  tap - Ralph Loop (${MODE} mode)${NC}"
echo -e "${GREEN}  Max iterations: ${MAX_ITERATIONS}${NC}"
echo -e "${GREEN}═══════════════════════════════════════${NC}"

# Select prompt file
if [ "$MODE" = "plan" ]; then
    PROMPT_FILE="PROMPT_plan.md"
else
    PROMPT_FILE="PROMPT_build.md"
fi

if [ ! -f "$PROMPT_FILE" ]; then
    echo -e "${RED}Error: $PROMPT_FILE not found${NC}"
    exit 1
fi

# Circuit breaker - detect stagnation
LAST_COMMIT=""
STAGNANT_COUNT=0
MAX_STAGNANT=3

while [ $ITERATION -lt $MAX_ITERATIONS ]; do
    ITERATION=$((ITERATION + 1))
    
    echo ""
    echo -e "${YELLOW}─── Iteration $ITERATION/$MAX_ITERATIONS ───${NC}"
    echo ""
    
    # Run Claude Code with the prompt
    cat "$PROMPT_FILE" | claude --dangerously-skip-permissions
    
    # Check for progress (new commits)
    CURRENT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "none")
    
    if [ "$CURRENT_COMMIT" = "$LAST_COMMIT" ]; then
        STAGNANT_COUNT=$((STAGNANT_COUNT + 1))
        echo -e "${YELLOW}No new commits (stagnant: $STAGNANT_COUNT/$MAX_STAGNANT)${NC}"
        
        if [ $STAGNANT_COUNT -ge $MAX_STAGNANT ]; then
            echo -e "${RED}Circuit breaker: $MAX_STAGNANT iterations without progress${NC}"
            echo -e "${YELLOW}Review @fix_plan.md and specs, then restart${NC}"
            exit 1
        fi
    else
        STAGNANT_COUNT=0
        LAST_COMMIT="$CURRENT_COMMIT"
        echo -e "${GREEN}✓ Progress detected${NC}"
    fi
    
    # Check if plan is empty (all done)
    if [ -f "@fix_plan.md" ]; then
        REMAINING=$(grep -c "^\- \[ \]" @fix_plan.md 2>/dev/null || echo "0")
        echo -e "Remaining tasks: $REMAINING"
        
        if [ "$REMAINING" = "0" ]; then
            echo -e "${GREEN}═══════════════════════════════════════${NC}"
            echo -e "${GREEN}  All tasks complete!${NC}"
            echo -e "${GREEN}═══════════════════════════════════════${NC}"
            exit 0
        fi
    fi
    
    # Small delay to allow review of output
    sleep 2
done

echo -e "${YELLOW}Reached max iterations ($MAX_ITERATIONS)${NC}"
echo -e "${YELLOW}Run again to continue, or review progress${NC}"
