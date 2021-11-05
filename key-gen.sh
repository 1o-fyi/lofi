#!/bin/bash

# First argument of script is your username
_NAME=${1:-"None"} && if [ "$_NAME" == "None" ]; then echo "pls enter name" && exit 1; fi
umask 077
cat <(age-keygen) | tail -n 1 | tr -d '\n' > "$_NAME.age"

