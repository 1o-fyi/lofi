#!/bin/bash

# First argument of script is your username
_NAME=${1:-"None"} && if [ "$_NAME" == "None" ]; then echo -e "usage: \n\t./keygen.sh [username]" && exit 1; fi
umask 077

# if no local build, then build.
if ! [ -f "lofi" ]; then
    make
fi

# safety check to prevent overwriting your keys
if [ -f "$_NAME.key" ] || [ -f "$_NAME.pub" ]; then
    read -e -p "keys for this username exist, are you sure you want to overwrite them? [ Y/n ]: " __OWRITE
    if [ "$__OWRITE" != "Y" ]; then
        echo -e "\nexiting with no overwrite.\n"
        exit 0
    fi
fi

# write public key to _NAME.age.pub
# write private key to _NAME.age
cat <(age-keygen 2> >(tee "$_NAME-age.pub" >&2)) | tail -n 1 | tr -d '\n' > "$_NAME.key"

./lofi fmt -P "$_NAME.key" -U "$_NAME" | tail -n 2 | tr -d '\n' > "$_NAME.pub"

echo -e "-  -  -"
echo "welcome $_NAME so couple of things to note"
echo "./$_NAME.key is your private key, keep it safe & back it up."
echo "./$_NAME.pub is your public key, you can lose this, share it with friends, or even derive it from your username."
echo "your username will be $_NAME, you can rerun this script to change this or if the name is already taken"
echo -e "- - -"
echo -e "\n\nnow to register open an issue here with the following: \n\n\thttps://github.com/1o-fyi/register/issues/new"
echo -e "\n\`\`\`"
cat "$_NAME.pub"
echo -e "\n\`\`\`\n"
echo -e "to see an example issue check out: https://github.com/1o-fyi/register/issues/2"
echo -e "if we don't know eachother that's cool just say hi and why you'd like to register, else you can just share the public key"
