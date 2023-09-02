#!/bin/bash

# # Start the Tailwind CSS watcher in the background
# pnpm start 2> tailwind.log&

# Start the Go server using air
op run --env-file .env -- air -c air.toml

