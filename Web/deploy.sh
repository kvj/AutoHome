#!/bin/sh

git pull origin master
node_modules/.bin/gulp dist
sudo service autohome-web restart
