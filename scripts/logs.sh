#!/usr/bin/env sh

find log -name "*.log" | while read -r l; do
  echo "\033[1;33m$l\033[0m"
  cat "$l"
  echo
done
