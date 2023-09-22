package repository

import "testing"

func TestNewZapLogger(t *testing.T) {
	l := NewZapLogger()
	l.Infof("asdfasdf %s", "aaa")
}
