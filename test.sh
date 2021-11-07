#!/bin/bash
# setup
# make install-scripts
# sudo ./install-go
readonly _version="$(./tag)"
readonly _uuid="gh-actions-$_version-$RANDOM"
readonly _api=${1:-"https://dev.1o.fyi"}

echo "running build $_uuid"
__EXIT_CODE=0

function clean
{
    rm -rf *.age.pub *.age
}

# compile
make

# setup alice, bob keys
_alice="alice"
_bob="bob"
_alice_pb="$_alice".age.pub
_bob_pb="$_bob".age.pub
_alice_msg="$(echo hi from alice | base64)"
_bob_msg="$(echo hello from bob | base64)"

# generate keys
./key-gen.sh "$_alice"
./key-gen.sh "$_bob"

# cat the public keys
_alice_pubkey="$(cat $_alice_pb | sed 's/Public key: //g' | tr -d '\n')"
_bob_pubkey="$(cat $_bob_pb | sed 's/Public key: //g' | tr -d '\n')"

# register public keys
echo "curl $_api/set?$_alice=$_alice_pubkey" | bash
echo "curl $_api/set?$_bob=$_bob_pubkey" | bash

echo -e "\nsetup: sending messages\n"
_alice_recv_cmd="$(./lofi -q -A $_api s -m $_bob_msg -r $_alice | tail -n 2 | tr -d '\n\t')"
_bob_recv_cmd="$(./lofi -q -A $_api s -m $_alice_msg -r $_bob | tail -n 2 | tr -d '\n\t')"

echo -e "\nsetup: receiving messages\n"
_alice_recv="$(echo ./lofi -U alice -A $_api r -k $_alice_recv_cmd -p alice.age | bash | tail -n 1 | tr -d '\n\t')"
_bob_recv="$(echo ./lofi -U bob -A $_api r -k $_bob_recv_cmd -p bob.age | bash | tail -n 1 | tr -d '\n\t')"

echo " - - - "

echo -e "\nchecking that bob received alices message"
if [ "$_alice_recv" != "$_bob_msg" ]; then 
    echo -e "\n- - fail - -\nalice received incorrect msg: \ngot\n\t$_alice_recv\nwanted\n\t$_bob_msg\n- - -\n"    
    __EXIT_CODE=1
else 
    echo -e "\n- - success! - -\nalice got \n\t$_alice_recv\n"
fi

echo -e "\nchecking that bob received alices message"
if [ "$_bob_recv" != "$_alice_msg" ]; then 
    echo -e "\n- - FAIL - -\nbob received incorrect msg: \ngot\n\t$_bob_recv\nwanted\n\t$_alice_msg\n- - -\n"
    __EXIT_CODE=1
else
    echo -e "\n- - success! - -\nbob got \n\t$_bob_recv\n"
fi

clean
echo -e "\n- - - Finished with exit code $__EXIT_CODE - - -\n"
exit $__EXIT_CODE
