#!/bin/bash
set -e

# Check if Oryx is responding
if curl -sf http://localhost:2022/terraform/v1/ffmpeg/query > /dev/null; then
  echo "Oryx is healthy"
  exit 0
else
  echo "Oryx is not responding"
  exit 1
fi


