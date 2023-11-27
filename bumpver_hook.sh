#!/bin/sh
git cliff -c cliff.yaml --tag 3.0.0 -o CHANGELOG.md
git add CHANGELOG.md