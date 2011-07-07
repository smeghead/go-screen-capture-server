#!/bin/bash
DISPLAY_NO=$1
URL=$2

DIR=$(dirname $0)
#IMAGE_FILE="${DIR}/images/tmp_${DISPLAY_NO}"

echo "display: $DISPLAY_NO"
echo "url: $URL"

#rm -rf /home/smeghead/.mozilla/C
echo firefox -display :$DISPLAY_NO -remote "openUrl($URL)"
firefox -display :$DISPLAY_NO -remote "openUrl($URL)"
#sleep 3
#WINID=`xwininfo -display :$DISPLAY_NO.0 -root -tree | grep "Mozilla Firefox" | head -n 1 | sed -e 's/^[  ]*//' | cut -f 1 -d ' '`
#echo "WINID: $WINID"
#xwd -display :1.0 -id $WINID -out ${IMAGE_FILE}.xwd
#convert ${IMAGE_FILE}.xwd ${IMAGE_FILE}.png
