#!/bin/bash

# try to figure out where we're running from, then delete contents of scratch dir
if [ -d "test/scratch" ]; then
  echo -e "deleting test scratch files:\n"
  rm -rfv test/scratch/*
elif [ -d "../scratch" ]; then
  echo -e "deleting test scratch files:\n"
  rm -rfv ../scratch/*
else
  echo "Don't know where test scratch dir is."
  echo -e "Run from veb dir or test/scripts dir.\n"
  exit 1
fi
