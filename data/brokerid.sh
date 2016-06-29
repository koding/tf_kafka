#!/bin/bash

FOURTHPOWER=`echo '256^3' | bc`
THIRDPOWER=`echo '256^2' | bc`
SECONDPOWER=`echo '256^1' | bc`

ID=`curl http://169.254.169.254/latest/meta-data/local-ipv4`
FOURTHIP=`echo $ID | cut -d '.' -f 1`
THIRDIP=`echo $ID | cut -d '.' -f 2`
SECONDIP=`echo $ID | cut -d '.' -f 3`
FIRSTIP=`echo $ID | cut -d '.' -f 4`
BROKER_ID=`expr $FOURTHIP \* $FOURTHPOWER + $THIRDIP \* $THIRDPOWER + $SECONDIP \* $SECONDPOWER + $FIRSTIP`
echo $BROKER_ID
