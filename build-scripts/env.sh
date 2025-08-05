#!/bin/bash

VENV_DIR="$HOME/.virtualenvs"
DOCS_VENV_DIR="$VENV_DIR/tupi-docs"
install_courtney(){
    go get golang.org/x/tools@v0.25.0
    go get github.com/dave/courtney@v0.4.3
    go install github.com/dave/courtney@v0.4.3
}

setup_env(){
    install_courtney
}

setup_docs_env(){
if [ ! -d "$DOCS_VENV_DIR" ]
    then
        echo "creating venv at $DOCS_VENV_DIR"
        mkdir -p $VENV_DIR
        python3 -m venv $DOCS_VENV_DIR
    fi
    source $DOCS_VENV_DIR/bin/activate
    echo "installing sphinx"
    pip install sphinx sphinx-pdj-theme
}

build_docs(){
    source $DOCS_VENV_DIR/bin/activate
    cd docs
    make html
    cd ..
}


case "$1" in
    "setup-env")
        setup_env
        ;;

    "setup-docs-env")
	setup_docs_env
	;;

    "build-docs")
        build_docs
        ;;

    *)

        echo "Usage: env.sh OP"
        echo "OPs are:"
        echo " - setup-env"
        echo " - build-docs"
        exit 1;

esac
