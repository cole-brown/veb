#!/bin/bash

# vars
SCRATCH_DIR="foo"
DATA_DIR="bar"

# try to figure out where we're running from, then delete contents of scratch dir
if [ -d "test/scratch" ]; then
  SCRATCH_DIR="test/scratch"
  DATA_DIR="test/data"
elif [ -d "../scratch" ]; then
  SCRATCH_DIR="../scratch"
  DATA_DIR="../data"
else
  echo -e "Don't know where test dirs are."
  echo -e "Run from veb dir or test/scripts dir.\n"
  exit 1
fi

FILES=${DATA_DIR}/small-files-change.txt

echo "Changing files!"
cat $FILES | tr "\n" "\0" | xargs -0 -I{} sh -c 'echo foo >> $1/$2' -- $SCRATCH_DIR {}
echo "  done."