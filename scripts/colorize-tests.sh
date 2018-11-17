#!/usr/bin/env bash

if (( $# == 0 )) ; then
    sed ''/PASS/s//$(printf "\033[32mPASS\033[0m")/'' < /dev/stdin | sed ''/FAIL/s//$(printf "\033[31mFAIL\033[0m")/''
else
    sed ''/PASS/s//$(printf "\033[32mPASS\033[0m")/'' <<< "$1" | sed ''/FAIL/s//$(printf "\033[31mFAIL\033[0m")/''
fi

echo "Code that cannot be tested is flawed :-)"