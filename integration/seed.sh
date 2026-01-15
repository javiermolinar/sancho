#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export DEEPWORK_DB_PATH="$SCRIPT_DIR/testdata/seed.db"
DW="./bin/sancho"

rm -f "$DEEPWORK_DB_PATH"

# Monday 2025-01-13
$DW add "Daily standup" --date=2025-01-13 --start=09:00 --end=09:30 --category=shallow
$DW add "Implement user authentication" --date=2025-01-13 --start=09:30 --end=12:00 --category=deep
$DW add "Email and Slack catchup" --date=2025-01-13 --start=13:00 --end=13:30 --category=shallow
$DW add "Database schema design" --date=2025-01-13 --start=13:30 --end=17:00 --category=deep

# Tuesday 2025-01-14
$DW add "Daily standup" --date=2025-01-14 --start=09:00 --end=09:30 --category=shallow
$DW add "Build REST endpoints" --date=2025-01-14 --start=09:30 --end=12:00 --category=deep
$DW add "Code review PRs" --date=2025-01-14 --start=13:00 --end=14:00 --category=shallow
$DW add "Frontend integration" --date=2025-01-14 --start=14:00 --end=17:00 --category=deep

# Wednesday 2025-01-15
$DW add "Daily standup" --date=2025-01-15 --start=09:00 --end=09:30 --category=shallow
$DW add "Write unit tests" --date=2025-01-15 --start=09:30 --end=12:00 --category=deep
$DW add "Team sync meeting" --date=2025-01-15 --start=13:00 --end=14:00 --category=shallow
$DW add "Refactor authentication module" --date=2025-01-15 --start=14:00 --end=17:00 --category=deep

# Thursday 2025-01-16
$DW add "Daily standup" --date=2025-01-16 --start=09:00 --end=09:30 --category=shallow
$DW add "API documentation" --date=2025-01-16 --start=09:30 --end=11:30 --category=deep
$DW add "1:1 with manager" --date=2025-01-16 --start=11:30 --end=12:00 --category=shallow
$DW add "Performance optimization" --date=2025-01-16 --start=13:00 --end=17:00 --category=deep

# Friday 2025-01-17
$DW add "Daily standup" --date=2025-01-17 --start=09:00 --end=09:30 --category=shallow
$DW add "Bug fixes" --date=2025-01-17 --start=09:30 --end=12:00 --category=deep
$DW add "Sprint retrospective" --date=2025-01-17 --start=13:00 --end=14:00 --category=shallow
$DW add "Sprint planning" --date=2025-01-17 --start=14:00 --end=15:30 --category=shallow
$DW add "Technical debt cleanup" --date=2025-01-17 --start=15:30 --end=17:00 --category=deep

# Saturday 2025-01-18
$DW add "Read technical book" --date=2025-01-18 --start=10:00 --end=12:00 --category=deep
$DW add "Side project hacking" --date=2025-01-18 --start=14:00 --end=17:00 --category=deep

# Sunday 2025-01-19
$DW add "Week planning" --date=2025-01-19 --start=10:00 --end=11:00 --category=shallow
$DW add "Learning new framework" --date=2025-01-19 --start=14:00 --end=16:00 --category=deep

echo "Created seed.db at $DEEPWORK_DB_PATH"
