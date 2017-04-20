#!/bin/bash

if [ -n "$(gofmt -l .)" ]; then
    echo "Go code is not formatted:"
    gofmt -d -e .
    exit 1
fi
