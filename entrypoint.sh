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

clair-action report \
    --image-path=${GITHUB_WORKSPACE}/${imagePath} \
    --image-ref=${imageRef} \
    --format=${format} > ${output} \
    --db-url=https://clair-sqlite-db.s3.amazonaws.com/matcher.zst

exit $returnCode
