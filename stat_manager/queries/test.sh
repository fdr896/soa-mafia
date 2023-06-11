#!/bin/bash

./query_player.sh test_user1
echo
./create.sh test_user1
echo
./query_player.sh test_user1
echo

./create_with_avatar.sh test_user2
echo
./create.sh test_user3
echo
./create_with_avatar.sh test_user4
echo

./query_players.sh test_user1,test_user3
echo
./query_players.sh test_user1,test_user2,test_user3,test_user4
echo
./query_players.sh test_user3
echo
./query_players.sh test_user
echo

./update_all.sh test_user
./update_all.sh test_user1
./update_avatar.sh test_user2
./update_email.sh test_user3
./update_email_gender.sh test_user4
echo

echo
./query_player.sh test_user1
echo
./query_player.sh test_user2
echo
./query_player.sh test_user3
echo
./query_player.sh test_user4
echo

./delete_player.sh test_user1
./delete_player.sh test_user2
./delete_player.sh test_user3
./delete_player.sh test_user4
./delete_player.sh test_user4
