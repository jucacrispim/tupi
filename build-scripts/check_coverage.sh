#!/bin/bash

courtney_cmd=courtney
if [ "$ENV" == "ci" ]
then
   courtney_cmd=~/go/bin/courtney
fi
out=`$courtney_cmd -e -v -t ./...`
error=$?

if [ "$error" != "0" ]
then
    echo ""
    echo "Something went wrong"
    echo "$out"
    exit 1
else
    echo ""
    echo "Yay! Everthing ok!"
fi
