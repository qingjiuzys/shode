#!/bin/sh
# Cron scheduler implementation

# Job storage
declare -A CRON_JOBS
declare -A CRON_COMMANDS
CRON_RUNNING=false

# CronSchedule schedules a new cron job
CronSchedule() {
    local schedule="$1"
    local command="$2"
    local job_name="${3:-job_$(date +%s)}"
    
    # Validate cron expression (basic validation)
    if ! _validate_cron "$schedule"; then
        echo "Error: Invalid cron expression: $schedule" >&2
        return 1
    fi
    
    CRON_JOBS["$job_name"]="$schedule"
    CRON_COMMANDS["$job_name"]="$command"
    
    echo "Scheduled job '$job_name': $schedule -> $command"
}

# _validate_cron validates a cron expression
_validate_cron() {
    local expr="$1"
    
    # Basic validation: 5 parts separated by spaces or asterisks
    local parts
    parts=($(echo "$expr" | tr ' ' '\n'))
    
    if [ "${#parts[@]}" -ne 5 ]; then
        return 1
    fi
    
    return 0
}

# CronStart starts the cron scheduler
CronStart() {
    if [ "$CRON_RUNNING" = "true" ]; then
        echo "Cron scheduler is already running" >&2
        return 1
    fi
    
    CRON_RUNNING=true
    echo "Cron scheduler started"
    
    # Main scheduler loop
    while [ "$CRON_RUNNING" = "true" ]; do
        local now
        now=$(date +%M:%H:%d:%m:%w)
        
        for job_name in "${!CRON_JOBS[@]}"; do
            local schedule="${CRON_JOBS[$job_name]}"
            local command="${CRON_COMMANDS[$job_name]}"
            
            if _should_run "$schedule" "$now"; then
                echo "Running job: $job_name"
                eval "$command" &
            fi
        done
        
        sleep 60
    done
}

# _should_run checks if a job should run at current time
_should_run() {
    local schedule="$1"
    local current="$2"
    
    # Simplified cron matching
    # Full implementation would parse cron expressions properly
    return 0
}

# CronStop stops the cron scheduler
CronStop() {
    if [ "$CRON_RUNNING" != "true" ]; then
        echo "Cron scheduler is not running" >&2
        return 1
    fi
    
    CRON_RUNNING=false
    echo "Cron scheduler stopped"
}

# CronList lists all scheduled jobs
CronList() {
    if [ "${#CRON_JOBS[@]}" -eq 0 ]; then
        echo "No jobs scheduled"
        return 0
    fi
    
    echo "Scheduled jobs:"
    for job_name in "${!CRON_JOBS[@]}"; do
        echo "  $job_name: ${CRON_JOBS[$job_name]}"
        echo "    Command: ${CRON_COMMANDS[$job_name]}"
    done
}
