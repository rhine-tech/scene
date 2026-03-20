package cache

import "time"

type Metrics interface {
	RecordGet(hit bool, d time.Duration, err error)
	RecordLoad(d time.Duration, err error)
	RecordSet(d time.Duration, err error)
	RecordDelete(d time.Duration, err error)
	RecordDecodeError(err error)
}

type NopMetrics struct{}

func (NopMetrics) RecordGet(bool, time.Duration, error) {}
func (NopMetrics) RecordLoad(time.Duration, error)      {}
func (NopMetrics) RecordSet(time.Duration, error)       {}
func (NopMetrics) RecordDelete(time.Duration, error)    {}
func (NopMetrics) RecordDecodeError(error)              {}
