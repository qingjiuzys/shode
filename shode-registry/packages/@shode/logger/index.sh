#!/bin/sh
# @shode/logger - Structured logging library for Shode applications

# Source dependencies
. "$(dirname "$0")/src/logger.sh"

# Export public API
export LogInfo
export LogWarn
export LogError
export LogDebug
export SetLogLevel
export AddLogTransport
