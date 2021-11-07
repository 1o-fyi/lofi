#!/bin/bash

readonly _version="$(./tag)"
readonly _uuid="gh-actions-$_version-$RANDOM"
readonly _api=${1:-"https://dev.1o.fyi"}
echo "running build $_uuid"

function clean
{
    rm -rf *.age.pub *.age result
}

function _test
{
    local _got_msg="$1"
    local _want_msg="$2"
    local _sender="$3"
    local _receiver="$4"
    local _exit_code="$(cat result)"
    echo -e "\nchecking that $_sender received $_receiver's message"
    if [ "$_got_msg" != "$_want_msg" ]; then 
        echo -e "\n- - failure - -\n $_sender received incorrect msg:\n\tgot:\n\t\t$_alice_recv\n\twanted\n\t\t$_bob_msg"    
        _exit_code=1
    else 
        echo -e "\n- - success! - -\n"
    fi

    echo $_exit_code > result &
}

# setup names, filepaths & messages
readonly _alice="alice"
readonly _bob="bob"
readonly _alice_pb="$_alice".age.pub
readonly _bob_pb="$_bob".age.pub
readonly _alice_msg="$(echo hi from alice | base64)"
readonly _bob_msg="$(echo hello from bob | base64)"

# setup a fifo queue for storing the exit code
__EXIT_CODE=0 && mkfifo result && echo $__EXIT_CODE > result & 

# build locally if none exists
if ! [ -f "./lofi" ]; then echo "no local build" && make; fi

# generate keys
./key-gen.sh "$_alice"
./key-gen.sh "$_bob"

# cat the public keys
_alice_pubkey="$(cat $_alice_pb | sed 's/Public key: //g' | tr -d '\n')"
_bob_pubkey="$(cat $_bob_pb | sed 's/Public key: //g' | tr -d '\n')" 

echo -e "\nsetup: setting public keys\n"
echo "curl $_api/set?$_alice=$_alice_pubkey" | bash && echo
echo "curl $_api/set?$_bob=$_bob_pubkey" | bash && echo

echo -e "\nsetup: sending messages\n"
_alice_recv_cmd="$(./lofi -q -A $_api s -m $_bob_msg -r $_alice | tail -n 2 | tr -d '\n\t')"
_bob_recv_cmd="$(./lofi -q -A $_api s -m $_alice_msg -r $_bob | tail -n 2 | tr -d '\n\t')"

echo -e "\nsetup: receiving messages\n"
_alice_recv="$(echo ./lofi -U alice -A $_api r -k $_alice_recv_cmd -p alice.age | bash | tail -n 1 | tr -d '\n\t')"
_bob_recv="$(echo ./lofi -U bob -A $_api r -k $_bob_recv_cmd -p bob.age | bash | tail -n 1 | tr -d '\n\t')"

echo " - - - "

# run tests
_test "$_bob_recv" "$_alice_msg" "$_alice" "$_bob"
_test "$_alice_recv" "$_bob_msg" "$_bob" "$_alice"

# empty fifo queue, clean up keys & exit 0 if passed 1 if failed.
__EXIT_CODE="$(cat result)"
clean
echo -e "\n- - - Finished with exit code $__EXIT_CODE - - -\n"
exit $__EXIT_CODE
