#!/bin/bash

out=`courtney -e -v -t ./...`
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
