#!/bin/bash

cd $(dirname "${BASH_SOURCE[0]}")/../..

set -ex

psql -d nxpkg-test-db  -c 'drop schema public cascade; create schema public;'
