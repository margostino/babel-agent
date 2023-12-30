#!/bin/bash

ssh-add ~/.ssh/id_ed25519

# while [ -z "$SSH_AUTH_SOCK" ]; do
#     sleep 1
# done

/Users/margostino/workspace/babel-agent/babel-agent --config /Users/margostino/.babel/babel.toml
