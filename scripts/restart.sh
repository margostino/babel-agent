#!/bin/bash

# Get the PID of the daemon 'org.babel.agent'
pid=$(sudo launchctl list | grep 'org.babel.agent' | awk '{print $1}')

# Check if the PID was found
if [ -z "$pid" ]; then
  echo "The daemon 'org.babel.agent' is not running."
else
  echo "Killing 'org.babel.agent' with PID $pid."
  # Kill the process with PID
  kill -9 $pid
fi
