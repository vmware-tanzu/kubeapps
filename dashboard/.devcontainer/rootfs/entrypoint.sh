#!/bin/bash

export PATH="/app/node_modules/.bin:$PATH"

echo "Node Version:  $(node -v)"

if ! [ -v SKIP_YARN_DEP_CHECK ]; then
  echo "Checking dependencies..."
  if ! yarn check > /dev/null 2>&1; then
    echo "Updating dependencies..."
    yarn install
  else
    echo "All dependencies are up to date"
  fi
fi

exec "$@"
