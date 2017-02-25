#!/bin/bash

# wrk must be available on PATH
# https://github.com/wg/wrk
echo "Dynamic test"
wrk -c100 -d10 -t10 http://127.0.0.1:8080/some/page/123

echo
echo "Static test"
wrk -c100 -d10 -t10 http://127.0.0.1:8080/other/page/path
