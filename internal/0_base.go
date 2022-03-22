package internal

import (
	// "fmt"
	"math/rand"
	"regexp"
	"time"
)

const (
	DefaultDir  = "temp"
	DefaultHead = `#! /usr/bin/env bash
set -eu -o pipefail
_wd=$(pwd)/
_path=$(dirname $0 | xargs -i readlink -f {})

`
)

var (
	_Rand        = rand.New(rand.NewSource(time.Now().UnixNano()))
	_LetterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	_TaskNameStr = `[0-9a-zA-Z_.-]{1,32}`
	_TaskNameRE  *regexp.Regexp
)

func init() {
	_TaskNameRE = regexp.MustCompile(_TaskNameStr)
}

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = _LetterRunes[_Rand.Intn(len(_LetterRunes))]
	}
	return string(b)
}

func Jobname(tn, object string) (now time.Time, name string) {
	now = time.Now()
	if object == "" {
		object = RandString(8)
	}
	name = tn + "__" + object + "." + now.Format("2006-01-02T15-04-05")
	return
}
