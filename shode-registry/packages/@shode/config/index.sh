#!/bin/sh
# @shode/config - Configuration management library

. "$(dirname "$0")/src/config.sh"

# Export public API
export ConfigLoad
export ConfigGet
export ConfigSet
export ConfigHas
export ConfigMerge
