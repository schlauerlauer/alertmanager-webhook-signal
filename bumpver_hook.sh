#!/bin/sh
git cliff -c cliff.yaml --tag 1.1.0 -o CHANGELOG.md
git add CHANGELOG.md