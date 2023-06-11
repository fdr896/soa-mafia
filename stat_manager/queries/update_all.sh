#!/bin/bash

curl -X PUT http://localhost:9077/player/$1 \
  -F "email=a@b.c" \
  -F "gender=male" \
  -F "avatar=@avatar.jpeg" \
  -H "Content-Type: multipart/form-data"
