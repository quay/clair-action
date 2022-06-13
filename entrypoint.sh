#!/bin/bash
set -e

while getopts "r:p:f:o:c:d:u:w:" o; do
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
       d)
         export dbURL="$(sed -e 's/^[ \t]*//'<<<"${OPTARG}")"
       ;;
       u)
         export dockerConfigDir="$(sed -e 's/^[ \t]*//'<<<"${OPTARG}")"
       ;;
  esac
done

if [[ -z "$dbURL" ]]; then
   dbURL="https://clair-sqlite-db.s3.amazonaws.com/matcher.zst"
fi

echo ${HOME}
ls -l ${HOME}
echo ${GITHUB_WORKSPACE}
ls -l ${GITHUB_WORKSPACE}

clair-action report \
    --image-path=${GITHUB_WORKSPACE}/${imagePath} \
    --image-ref=${imageRef} \
    --docker-config-dir=${GITHUB_WORKSPACE}/${dockerConfigDir} \
    --db-url=${dbURL} \
    --return-code=${returnCode} \
    --format=${format} > ${output}
