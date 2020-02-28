package motion

import "time"

// DelayedState - simple motion delayed state based on time
type DelayedState struct {
	lastOn  time.Time
	lastOff time.Time
	minOn   time.Duration
	minOff  time.Duration
	on      bool
}

// NewDelayedState creates a new motion delayed state DelayedState
func NewDelayedState(minOn time.Duration, minOff time.Duration) DelayedState {
	return DelayedState{
		minOn:  minOn,
		minOff: minOff,
		on:     false,
	}
}

func (md *DelayedState) detect(on bool, now time.Time) bool {
	if on {
		md.lastOn = now
	} else {
		md.lastOff = now
	}

	if md.lastOn == (time.Time{}) {
		md.lastOn = now
	}

	if md.lastOff == (time.Time{}) {
		md.lastOff = now
	}

	diff := md.lastOn.Sub(md.lastOff)
	if diff > md.minOn && md.on == false {
		md.on = true
	} else if (diff*-1) > md.minOff && md.on == true {
		md.on = false
	}

	return md.on
}