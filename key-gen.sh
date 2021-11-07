#!/bin/bash

# First argument of script is your username
_NAME=${1:-"None"} && if [ "$_NAME" == "None" ]; then echo "pls enter name" && exit 1; fi
umask 077
# write public key to _NAME.age.pub
# write private key to _NAME.age
cat <(age-keygen 2> >(tee "$_NAME.age.pub" >&2)) | tail -n 1 | tr -d '\n' > "$_NAME.age"