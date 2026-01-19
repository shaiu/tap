#!/bin/bash
# Ralph Wiggum loop for tap - INTERACTIVE MODE
# Usage: ./loop-interactive.sh [plan|build]

set -e

MODE="${1:-build}"
ITERATION=0

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${GREEN}═══════════════════════════════════════${NC}"
echo -e "${GREEN}  tap - Ralph Loop (${MODE} mode)${NC}"
echo -e "${GREEN}  INTERACTIVE - You approve each step${NC}"
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

while true; do
    ITERATION=$((ITERATION + 1))
    
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════${NC}"
    echo -e "${CYAN}  Iteration $ITERATION${NC}"
    echo -e "${CYAN}═══════════════════════════════════════${NC}"
    echo ""
    
    # Show current task
    if [ -f "@fix_plan.md" ]; then
        echo -e "${YELLOW}Current task:${NC}"
        grep -A 1 "← CURRENT" @fix_plan.md 2>/dev/null || grep "^\- \[ \]" @fix_plan.md | head -1
        echo ""
        
        REMAINING=$(grep -c "^\- \[ \]" @fix_plan.md 2>/dev/null || echo "0")
        echo -e "Remaining tasks: ${REMAINING}"
        echo ""
        
        if [ "$REMAINING" = "0" ]; then
            echo -e "${GREEN}═══════════════════════════════════════${NC}"
            echo -e "${GREEN}  All tasks complete! 🎉${NC}"
            echo -e "${GREEN}═══════════════════════════════════════${NC}"
            exit 0
        fi
    fi
    
    # Ask for approval before starting
    echo -e "${YELLOW}Ready to start iteration $ITERATION${NC}"
    echo ""
    echo "  [Enter]  Start this iteration"
    echo "  [s]      Skip to next task (mark current as done)"
    echo "  [v]      View the prompt that will be sent"
    echo "  [p]      View full task plan"
    echo "  [q]      Quit"
    echo ""
    read -p "Your choice: " choice
    
    case "$choice" in
        q|Q)
            echo -e "${YELLOW}Stopped at iteration $ITERATION${NC}"
            exit 0
            ;;
        s|S)
            echo -e "${YELLOW}Skipping... (manually mark task as done in @fix_plan.md)${NC}"
            ${EDITOR:-vim} @fix_plan.md
            continue
            ;;
        v|V)
            echo ""
            echo -e "${CYAN}─── Prompt Preview ───${NC}"
            cat "$PROMPT_FILE"
            echo -e "${CYAN}─── End Preview ───${NC}"
            echo ""
            read -p "Press Enter to continue..."
            continue
            ;;
        p|P)
            echo ""
            cat @fix_plan.md
            echo ""
            read -p "Press Enter to continue..."
            continue
            ;;
    esac
    
    # Run Claude Code WITHOUT --dangerously-skip-permissions
    # This means Claude will ask YOU for approval on each action
    echo ""
    echo -e "${GREEN}Starting Claude Code...${NC}"
    echo -e "${YELLOW}Claude will ask for your approval on file changes and commands${NC}"
    echo ""
    
    cat "$PROMPT_FILE" | claude
    
    # After Claude exits, show what happened
    echo ""
    echo -e "${CYAN}─── Iteration $ITERATION Complete ───${NC}"
    echo ""
    
    # Show recent commits
    echo -e "${YELLOW}Recent commits:${NC}"
    git log --oneline -3 2>/dev/null || echo "(no commits yet)"
    echo ""
    
    # Show changed files
    echo -e "${YELLOW}Uncommitted changes:${NC}"
    git status --short 2>/dev/null || echo "(not a git repo)"
    echo ""
    
    # Ask what to do next
    echo "What next?"
    echo "  [Enter]  Continue to next iteration"
    echo "  [r]      Retry this iteration"
    echo "  [e]      Edit task plan before continuing"
    echo "  [q]      Quit"
    echo ""
    read -p "Your choice: " next_choice
    
    case "$next_choice" in
        q|Q)
            echo -e "${YELLOW}Stopped after iteration $ITERATION${NC}"
            exit 0
            ;;
        r|R)
            ITERATION=$((ITERATION - 1))  # Will increment back at loop start
            echo -e "${YELLOW}Retrying iteration...${NC}"
            ;;
        e|E)
            ${EDITOR:-vim} @fix_plan.md
            ;;
    esac
done
