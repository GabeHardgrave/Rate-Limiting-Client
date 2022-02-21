#!/bin/bash

go test ./... \
  -v \
  -shuffle=on \
  -race \
  -covermode=atomic \
  -coverprofile=coverage.out \
