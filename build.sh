#!/bin/bash

mkdir -p output
rm -rf output/*
mkdir -p output/{gnlogd,gnlog-cli,gnlog-test}

cd gnlogd && go build -o ../output/gnlogd/gnlogd
[ $? -ne 0 ] && { echo "build 'gnlogd' failed"; exit 1; }
cd - >/dev/null

cd gnlog-cli && go build -o ../output/gnlog-cli/gnlog-cli
[ $? -ne 0 ] && { echo "build 'gnlog-cli' failed"; exit 1; }
cd - >/dev/null

cd gnlog-test && go build -o ../output/gnlog-test/gnlog-test
[ $? -ne 0 ] && { echo "build 'gnlog-test' failed"; exit 1; }
cd - >/dev/null

cp misc/* output/
cp -aR gnlogd/conf output/gnlogd/
cp -aR gnlogd/control.sh output/gnlogd/

cp -aR gnlog-cli/conf output/gnlog-cli/
cp -aR gnlog-cli/control.sh output/gnlog-cli/
