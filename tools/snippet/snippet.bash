#!/bin/bash

file=$1
pattern=$2

# read stdin
IN=$(cat -)

# write snippet to file so that `sed` can read
# it with `r` command (inserting directly is problematic
# when snippet contains newlines).
echo '```' >> snippettmp.txt
echo "$IN" >> snippettmp.txt
echo '```' >> snippettmp.txt

lead="<!-- BEGIN $pattern -->"
tail="<!-- END $pattern -->"

# replace snippet region with content from file
# this is taken from https://superuser.com/a/440057
out=$(sed -e "/$lead/,/$tail/{ /$lead/{p; r snippettmp.txt
 }; /$tail/p; d }" $file)

# write the file
echo "$out" > $file
rm snippettmp.txt

