#!/bin/bash

# Create 2GB files.

COUNTER=0
while [  $COUNTER -lt 2 ]; do
  echo "Creating file $COUNTER"
  openssl rand -out "testfile$COUNTER" -base64 $(( 2**30 * 3/4 * 2 ))
  let COUNTER=COUNTER+1 
done
