#!/bin/bash
set -e

local-clair report --db-path=/matcher $@

exit $returnCode
