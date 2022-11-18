#!/bin/bash
set -e

while getopts "r:p:f:o:c:d:u:w:b:" o; do
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
       w)
         export mode="$(sed -e 's/^[ \t]*//'<<<"${OPTARG}")"
       ;;
       b)
         export dbPath="$(sed -e 's/^[ \t]*//'<<<"${OPTARG}")"
       ;;
  esac
done

if [[ -z "$dbURL" ]]; then
   dbURL="https://clair-sqlite-db.s3.amazonaws.com/matcher.zst"
fi

if [[ ${mode} = "update" ]]
then
  clair-action update --db-path=${dbPath}
else
  clair-action report \
      --image-path=${GITHUB_WORKSPACE}/${imagePath} \
      --image-ref=${imageRef} \
      --docker-config-dir=${GITHUB_WORKSPACE}/${dockerConfigDir} \
      --db-url=${dbURL} \
      --db-path=${dbPath} \
      --return-code=${returnCode} \
      --format=${format} > ${output}
fi
