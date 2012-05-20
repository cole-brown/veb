#!/bin/bash

# vars
SCRATCH_DIR="foo"
DATA_DIR="bar"

# try to figure out where we're running from, then delete contents of scratch dir
if [ -d "test/scratch" ]; then
  SCRATCH_DIR="test/scratch/local"
  DATA_DIR="test/data"
elif [ -d "../scratch" ]; then
  SCRATCH_DIR="../scratch/local"
  DATA_DIR="../data"
else
  echo -e "Don't know where test dirs are."
  echo -e "Run from veb dir or test/scripts dir.\n"
  exit 1
fi

FILES=${DATA_DIR}/small-files.txt
DIRS=${DATA_DIR}/small-dirs.txt

echo "Making dirs..."
cat $DIRS | tr "\n" "\0" | xargs -0 -I{} mkdir -p "${SCRATCH_DIR}/{}"
echo "  done."

echo "Making files with random 2MB of content..."
cat $FILES | tr "\n" "\0" | xargs -0 -I{} sh -c 'dd if=/dev/urandom of="$1/$2" bs=1048576 count=2' -- $SCRATCH_DIR {}
echo "  done."