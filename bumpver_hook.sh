#!/bin/sh
git cliff -c cliff.yaml --tag 3.1.0 -o CHANGELOG.md
git add CHANGELOG.md