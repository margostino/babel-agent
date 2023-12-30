#!/bin/bash

# Wait for SSH_AUTH_SOCK to be set
while [ -z "$SSH_AUTH_SOCK" ]; do
    sleep 1
done

# Now start your daemon
/Users/margostino/workspace/babel-agent/babel-agent --config /Users/margostino/.babel/babel.toml
