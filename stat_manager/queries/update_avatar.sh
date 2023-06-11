#!/bin/bash

curl -X PUT http://localhost:9077/player/$1 \
  -F "avatar=@avatar.png" \
  -H "Content-Type: multipart/form-data"
