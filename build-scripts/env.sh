#!/bin/bash

install_courtney(){
    go get -u github.com/dave/courtney
    go install github.com/dave/courtney
    ls ~/go/bin
    ls ~/usr/go/bin
}

setup_env(){
    install_courtney
}


case "$1" in
    "setup-env")
        setup_env
        ;;

    *)

        echo "Usage: env.sh OP"
        echo "OPs are:"
        echo " - setup-env"
        exit 1;

esac
