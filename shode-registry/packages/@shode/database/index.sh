#!/bin/sh
# @shode/database - Database abstraction layer

. "$(dirname "$0")/src/database.sh"

# Export public API
export DbConnect
export DbQuery
export DbExec
export DbClose
export DbEscape
