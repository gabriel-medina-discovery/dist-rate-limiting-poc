#!/usr/bin/env sh

cd doc || exit
echo "Updating diagrams in doc folder ..."
plantuml -progress -r **/*.plantuml
echo
