package main

import (
	"regexp"
)

var g_isVarNameRegexp *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z_]\w*$`)
var g_isNumberRegexp *regexp.Regexp = regexp.MustCompile(`^(-)?[0-9]+$`)
var g_fetchListTypeRegexp *regexp.Regexp = regexp.MustCompile(`^list{(.+)}$`)
var g_notWordRegexp *regexp.Regexp = regexp.MustCompile(`[^\w]`)
