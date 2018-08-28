#!/bin/bash
rclone sync -v --delete-during --transfers 2 --checkers 4 --contimeout 60s --timeout 300s --retries 20 --low-level-retries 30 /home/jason/WKDownloader-NG/data hubic:LightNovel
