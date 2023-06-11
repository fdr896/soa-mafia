#!/bin/bash

curl -X POST http://localhost:9077/player \
  -F "username=$1" \
  -F "email=avatars_email@email.email" \
  -F "gender=male" \
  -F "avatar=@avatar.jpeg" \
  -H "Content-Type: multipart/form-data"
