#!/bin/bash

filename=$(basename "$0")
scriptname="${filename%.*}"
date "+${scriptname} start %Y%m%d_%H%M%S"

set -exo pipefail
export SHELLOPTS

WALLET=$1

gcmd="goal -w ${WALLET}"

ACCOUNT=$(${gcmd} account list|awk '{ print $3 }')

# Easier than prefixing all of the generated files.
cd "$TEMPDIR"

# The program the account will delegate to. It checks that it is being asked to
# authorize the transaction itself, which is what a delegated program approves.
cat > inner.teal <<EOF
#pragma version 14
arg 0
byte "inner"
==
global AuthMsg
txn TxID
==
&&
EOF

# A program's legacy address is its program hash in base32, which is exactly
# what global DelegatedProgramHash reports, so the delegator can name it.
INNER=$(${gcmd} clerk compile -n inner.teal | awk '{ print $2 }')

# The delegating account. It approves that one program and nothing else, and
# checks that it is approving a program rather than the transaction.
cat > deleg.teal <<EOF
#pragma version 14
arg 0
byte "delegator"
==
global DelegatedProgramHash
addr ${INNER}
==
&&
global AuthMsg
txn TxID
!=
&&
EOF

COMPILED=$(${gcmd} clerk compile -n deleg.teal)
DELEGATOR=$(echo "$COMPILED" | sed 's/.*(pq: \(.*\))/\1/')
echo "delegator account: $DELEGATOR"

# Fund the delegating account below one reward unit to avoid balance drift.
FUNDING=900000
${gcmd} clerk send -a "${FUNDING}" -f "${ACCOUNT}" -t "${DELEGATOR}"

DELEGATOR_ARG=$(printf 'delegator' | base64)
INNER_ARG=$(printf 'inner' | base64)

# Spending from the delegating account through the program it delegated to. No
# key is involved: the delegator approves the program, the program authorizes
# the transaction.
${gcmd} clerk send --delegator deleg.teal --delegator-argb64 "${DELEGATOR_ARG}" \
        -F inner.teal --argb64 "${INNER_ARG}" \
        -a 1000 -f "${DELEGATOR}" -t "${ACCOUNT}"

# The delegation names one program. Another, even one that approves on its own,
# is not what the delegating account agreed to.
printf '#pragma version 14\nint 1\n' > other.teal
set +o pipefail
${gcmd} clerk send --delegator deleg.teal --delegator-argb64 "${DELEGATOR_ARG}" \
        -F other.teal -a 1000 -f "${DELEGATOR}" -t "${ACCOUNT}" 2>&1 \
    | grep "rejected by delegating logic" || exit 1
set -o pipefail

# Each program reads its own arguments, so swapping them satisfies neither.
set +o pipefail
${gcmd} clerk send --delegator deleg.teal --delegator-argb64 "${INNER_ARG}" \
        -F inner.teal --argb64 "${DELEGATOR_ARG}" \
        -a 1000 -f "${DELEGATOR}" -t "${ACCOUNT}" 2>&1 \
    | grep "rejected by delegating logic" || exit 1
set -o pipefail

# Naming the delegating account is not required: with no --from, the delegator
# is the account the transaction comes from.
${gcmd} clerk send --delegator deleg.teal --delegator-argb64 "${DELEGATOR_ARG}" \
        -F inner.teal --argb64 "${INNER_ARG}" \
        -a 1000 -t "${ACCOUNT}"

BALANCE=$(${gcmd} account balance -a "${DELEGATOR}" | awk '{ print $1 }')
EXPECT=$((FUNDING - 2 * (1000 + 1000)))
if [ "$BALANCE" -ne "$EXPECT" ]; then
    date "+${scriptname} FAIL wanted balance=${EXPECT} but got ${BALANCE} %Y%m%d_%H%M%S"
    false
fi

date "+${scriptname} OK %Y%m%d_%H%M%S"
