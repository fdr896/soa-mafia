#!/bin/bash

curl -X PUT http://localhost:9077/player/$1 \
  -F "email=email@email.best" \
  -F "gender=female" \
  -H "Content-Type: multipart/form-data"
