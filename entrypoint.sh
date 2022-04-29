#!/bin/bash
set -e

while getopts "r:p:f:o:c:" o; do
   case "${o}" in
       r)
         export imageRef="$(sed -e 's/^[ \t]*//'<<<"${OPTARG}")"
       ;;
       p)
         export imagePath="$(sed -e 's/^[ \t]*//'<<<"${OPTARG}")"
       ;;
       f)
         export format="$(sed -e 's/^[ \t]*//'<<<"${OPTARG}")"
       ;;
       o)
         export output="$(sed -e 's/^[ \t]*//'<<<"${OPTARG}")"
       ;;
       c)
         export returnCode="$(sed -e 's/^[ \t]*//'<<<"${OPTARG}")"
       ;;
  esac
done

local-clair report \
    --db-path=/matcher \
    --image-path=${GITHUB_WORKSPACE}/${imagePath} \
    --image-ref=${imageRef} \
    --format=${format} > ${output}

exit $returnCode
