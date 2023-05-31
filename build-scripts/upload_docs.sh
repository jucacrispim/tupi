#!/bin/bash

cd docs/build
mv html tupi
tar -czf docs.tar.gz tupi

curl --user "$TUPI_USER:$TUPI_PASSWD" -F 'file=@docs.tar.gz' https://docs.poraodojuca.dev/e/
