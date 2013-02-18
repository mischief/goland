#!/bin/sh
repodir=golandsrc
exefile=goland
function stderr {
	echo "$@" 1>&2;
}
stderr "Cloning from github -> $repodir."
git clone git@github.com:mischief/goland.git $repodir > /dev/null
cd $repodir/cmd/goland
stderr "Getting deps, while helpless."
go get $(go list -f "{{range .Imports}}{{ . }} {{end}}" .)
stderr "Building goland, conducting genocide on cockatrices."
go build
stderr "DONE! type ./$exefile to play"
cd ../../../
ln -s ./$repodir/cmd/goland/goland $exefile
