#!/bin/sh

curl -F "file=@tester/test-upload.txt"  "http://127.0.0.1:2009/api/cdnize?api_key=123&file_name=test-upload.txt"

