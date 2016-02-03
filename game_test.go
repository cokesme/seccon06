package main

import (
	"testing"
	"time"
)

func TestQuestionIsOpen(t *testing.T) {
	q := &Question{
		hMap:     Map{},
		openTime: 30 * time.Second,
	}
	if q.IsOpen(time.Now(), time.Now().Add(10*time.Second)) != false {
		t.Error("error")
	}
	if q.IsOpen(time.Now(), time.Now().Add(40*time.Second)) != true {
		t.Error("error")
	}

}
