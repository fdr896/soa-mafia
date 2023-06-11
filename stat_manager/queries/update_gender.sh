#!/bin/bash

curl -X PUT http://localhost:9077/player/$1 \
  -F "gender=female" \
  -H "Content-Type: multipart/form-data"
