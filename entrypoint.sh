#!/bin/bash
set -e

while getopts "r:p:f:o:c:d:u:w:b:g:" o; do
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
       g)
         export clairConfig="$(sed -e 's/^[ \t]*//'<<<"${OPTARG}")"
       ;;
  esac
done

if [[ ${mode} = "update" ]]
then
  clair-action update --db-path=${dbPath} ${clairConfig:+--config=${GITHUB_WORKSPACE}/${clairConfig}}
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
