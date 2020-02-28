package motion

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDelayedState(t *testing.T) {
	minOn := 3 * time.Second
	minOff := 5 * time.Second
	s := NewDelayedState(minOn, minOff)
	assert.Equal(t, minOn, s.minOn)
	assert.Equal(t, minOff, s.minOff)
	assert.Equal(t, false, s.on)
}

func TestDetectStartFromOn(t *testing.T) {
	s := NewDelayedState(3*time.Second, 5*time.Second)
	tm := time.Date(2019, time.January, 21, 0, 0, 0, 0, time.UTC)
	r := s.detect(true, tm)
	assert.Equal(t, false, r)
	// 2
	tm = tm.Add(2 * time.Second)
	r = s.detect(true, tm)
	assert.Equal(t, false, r)
	// 4
	tm = tm.Add(2 * time.Second)
	r = s.detect(true, tm)
	assert.Equal(t, true, r)
	// 6
	tm = tm.Add(2 * time.Second)
	r = s.detect(false, tm)
	assert.Equal(t, true, r)
	// 8
	tm = tm.Add(2 * time.Second)
	r = s.detect(false, tm)
	assert.Equal(t, true, r)
	// 10
	tm = tm.Add(2 * time.Second)
	r = s.detect(false, tm)
	assert.Equal(t, false, r)
}

func TestDetectStartFromOff(t *testing.T) {
	s := NewDelayedState(3*time.Second, 5*time.Second)
	tm := time.Date(2019, time.January, 21, 0, 0, 0, 0, time.UTC)
	r := s.detect(false, tm)
	assert.Equal(t, false, r)
	// 2
	tm = tm.Add(2 * time.Second)
	r = s.detect(true, tm)
	assert.Equal(t, false, r)
	// 4
	tm = tm.Add(2 * time.Second)
	r = s.detect(true, tm)
	assert.Equal(t, true, r)
	// 6
	tm = tm.Add(2 * time.Second)
	r = s.detect(false, tm)
	assert.Equal(t, true, r)
	// 8
	tm = tm.Add(2 * time.Second)
	r = s.detect(false, tm)
	assert.Equal(t, true, r)
	// 10
	tm = tm.Add(2 * time.Second)
	r = s.detect(false, tm)
	assert.Equal(t, false, r)
}
