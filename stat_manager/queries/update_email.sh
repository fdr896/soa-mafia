#!/bin/bash

curl -X PUT http://localhost:9077/player/$1 \
  -F "email=new@email.email" \
  -H "Content-Type: multipart/form-data"
