#!/bin/bash
rclone sync -v --delete-during --transfers 2 --checkers 4 --contimeout 60s --timeout 300s --retries 10 --low-level-retries 15 /home/jason/WKDownloader-NG/data onedrive:Otaku/LightNovel/Books
