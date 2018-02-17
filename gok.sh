#!/bin/sh
# Copyright 2017 The Peggy Authors
#
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd.

#
# Verifies that go code passes go fmt, go vet, golint, and go test.
#

o=$(mktemp tmp.XXXXXXXXXX)

fail() {
	echo Failed
	cat $o
	rm $o
	exit 1
}

trap fail INT TERM

#echo Generating
#go generate . || fail

echo Formatting
gofmt -l $(find . -name '*.go') > $o 2>&1
test $(wc -l $o | awk '{ print $1 }') = "0" || fail

echo Vetting
go vet ./... > $o 2>&1 || fail

echo Testing
go test -test.timeout=60s ./... > $o 2>&1 || fail

echo Linting
golint ./... \
	| grep -v 'receiver name peggyrcvr should be consistent'\
	| grep -v 'const peggyEofCode should be peggyEOFCode'\
	| egrep -v 'grammar.y.*ALL_CAPS'\
	| egrep -v '(Begin|End|FullParenString|Type|CanFail).*should have comment or be unexported'\
	| egrep -v 'GenAccept should have comment or'\
	> $o 2>&1
# Silly: diff the grepped golint output with empty.
# If it's non-empty, error, otherwise succeed.
e=$(tempfile)
touch $e
diff $o $e > /dev/null || { rm $e; fail; }

rm $o $e
