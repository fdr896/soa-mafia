#!/bin/bash

curl -X POST http://localhost:9077/player \
  -F "username=$1" \
  -F "email=email@email.email" \
  -F "gender=male" \
  -H "Content-Type: multipart/form-data"
