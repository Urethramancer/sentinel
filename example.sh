#!/bin/bash
# This example script takes environment variables supplied by Sentinel
# and writes them to a log file with a time and date.

LOG="/var/log/sentinel.log"
date="date -Ins"

case "$SENTINEL_ACTION" in
	"create")
		echo `$date`: CREATE $SENTINEL_PATH >> $LOG
		;;
	"delete")
		echo `$date`: DELETE $SENTINEL_PATH >> $LOG
		;;
	"rename")
		echo `$date`: RENAME $SENTINEL_PATH >> $LOG
		;;
	"write")
		echo `$date`: WRITE $SENTINEL_PATH >> $LOG
		;;
	"chmod")
		echo `$date`: CHMOD $SENTINEL_PATH >> $LOG
		;;
esac
