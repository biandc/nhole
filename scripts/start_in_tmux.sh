#!/bin/bash
# Description:start nhole-client/nhole-server in tmux.
# @Author:biandc
# Parameters:nhole-client/nhole-server
set -e

test $# -eq 1 || exit

test "$1" = nhole-client || test "$1" = nhole-server || exit

tmux -V >/dev/null 2>&1 || exit

tmux has-session -t "$1" >/dev/null 2>&1 && exit

client() {
    chmod +x ./nhole-client
    [ -f "nhole-client.log" ] || touch nhole-client.log
    tmux new-session -s nhole-client -n start -d
    tmux send-keys -t nhole-client './nhole-client -c ./nhole-client.yaml --log_way="file" --log_file="./nhole-client.log" --log_level="debug"' C-m
    tmux new-window -n info -t nhole-client
    tmux send-keys -t nhole-client:1 'tail -f ./nhole-client.log' C-m
    tmux split-window -v -t nhole-client:1
    tmux send-keys -t nhole-client:1 'watch "netstat -antp | grep 127.0.0.1 | grep 22"' C-m
    return
}

server() {
    chmod +x ./nhole-server
    [ -f "nhole-server.log" ] || touch nhole-server.log
    tmux new-session -s nhole-server -n start -d
    tmux send-keys -t nhole-server './nhole-server -c ./nhole-server.yaml --log_way="file" --log_file="./nhole-server.log" --log_level="debug"' C-m
    tmux new-window -n info -t nhole-server
    tmux send-keys -t nhole-server:1 'tail -f ./nhole-server.log' C-m
    tmux split-window -v -t nhole-server:1
    tmux send-keys -t nhole-server:1 'watch "netstat -antp | grep 6553"' C-m
    return
}

case $1 in
nhole-client)
    echo "Start nhole-client ..."
    client
    echo "Successfully started nhole-client."
    ;;
nhole-server)
    echo "Start nhole-server ..."
    server
    echo "Successfully started nhole-server."
    ;;
*) ;;

esac
