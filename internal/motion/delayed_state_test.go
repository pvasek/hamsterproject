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
	ch, r := s.Detect(true, tm)
	assert.Equal(t, false, r)
	assert.Equal(t, false, ch)
	// 2
	tm = tm.Add(2 * time.Second)
	ch, r = s.Detect(true, tm)
	assert.Equal(t, false, r)
	assert.Equal(t, false, ch)
	// 4
	tm = tm.Add(2 * time.Second)
	ch, r = s.Detect(true, tm)
	assert.Equal(t, true, r)
	assert.Equal(t, true, ch)
	// 6
	tm = tm.Add(2 * time.Second)
	ch, r = s.Detect(false, tm)
	assert.Equal(t, true, r)
	assert.Equal(t, false, ch)
	// 8
	tm = tm.Add(2 * time.Second)
	ch, r = s.Detect(false, tm)
	assert.Equal(t, true, r)
	assert.Equal(t, false, ch)
	// 10
	tm = tm.Add(2 * time.Second)
	ch, r = s.Detect(false, tm)
	assert.Equal(t, false, r)
	assert.Equal(t, true, ch)
}

func TestDetectStartFromOff(t *testing.T) {
	s := NewDelayedState(3*time.Second, 5*time.Second)
	tm := time.Date(2019, time.January, 21, 0, 0, 0, 0, time.UTC)
	ch, r := s.Detect(false, tm)
	assert.Equal(t, false, r)
	assert.Equal(t, false, ch)
	// 2
	tm = tm.Add(2 * time.Second)
	ch, r = s.Detect(true, tm)
	assert.Equal(t, false, r)
	assert.Equal(t, false, ch)
	// 4
	tm = tm.Add(2 * time.Second)
	ch, r = s.Detect(true, tm)
	assert.Equal(t, true, r)
	assert.Equal(t, true, ch)
	// 6
	tm = tm.Add(2 * time.Second)
	ch, r = s.Detect(false, tm)
	assert.Equal(t, true, r)
	assert.Equal(t, false, ch)
	// 8
	tm = tm.Add(2 * time.Second)
	ch, r = s.Detect(false, tm)
	assert.Equal(t, true, r)
	assert.Equal(t, false, ch)
	// 10
	tm = tm.Add(2 * time.Second)
	ch, r = s.Detect(false, tm)
	assert.Equal(t, false, r)
	assert.Equal(t, true, ch)
}
