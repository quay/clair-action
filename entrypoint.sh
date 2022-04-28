#!/bin/bash
set -e

local-clair report --db-path=/matcher ${@:2} > $1

exit $returnCode
