#!/bin/bash

cd docs/build
mv html tupi
tar -czf docs.tar.gz tupi

curl -F 'file=@docs.tar.gz' https://docs.poraodojuca.dev/e/ -H "Authorization: Key $TUPI_AUTH_KEY"
