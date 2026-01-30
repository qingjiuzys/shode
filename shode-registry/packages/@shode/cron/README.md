# @shode/cron

Cron-like task scheduling for Shode applications.

## Features

- Schedule tasks using cron expressions
- Start/stop scheduler
- List all scheduled jobs
- Background job execution

## Installation

```bash
shode pkg add @shode/cron ^1.0.0
```

## Usage

```bash
. sh_modules/@shode/cron/index.sh

# Schedule a job to run every hour
CronSchedule "0 * * * *" "run_backup.sh" "hourly_backup"

# Schedule a job to run every day at midnight
CronSchedule "0 0 * * *" "cleanup.sh" "daily_cleanup"

# List all jobs
CronList

# Start scheduler (runs in background)
CronStart &

# Stop scheduler
CronStop
```

## API

### Functions

- `CronSchedule(schedule, command, name)` - Schedule a new job
- `CronStart()` - Start the scheduler
- `CronStop()` - Stop the scheduler
- `CronList()` - List all scheduled jobs

## Cron Expressions

```
* * * * *
│ │ │ │ │
│ │ │ │ └─ Day of week (0-6, 0 = Sunday)
│ │ │ └─── Month (1-12)
│ │ └───── Day of month (1-31)
│ └─────── Hour (0-23)
└───────── Minute (0-59)
```

## License

MIT
