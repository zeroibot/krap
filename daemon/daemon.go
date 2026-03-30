package daemon

import (
	"fmt"
	"time"

	"github.com/zeroibot/fn/clock"
	"github.com/zeroibot/fn/dict"
	"github.com/zeroibot/fn/io"
)

/*
Note: DaemonConfig is expected to have this structure:
Use int for Interval and Margin values so that we can disable
a daemon by setting the value to -1 or any value < 0

type Config sturct {
	<Domain> struct {
		<FeatureInterval> int
		...
	}
	...
}
*/

var (
	daemonStart    = dict.NewSyncMap[string, time.Time]()
	daemonLast     = dict.NewSyncMap[string, time.Time]()
	daemonDuration = dict.NewSyncMap[string, time.Duration]()
)

// Load Daemon Config which follows the expected structure,
// Validates if any of the Interval values are 0 (invalid)
func LoadConfig[T any](path string) (*T, error) {
	cfg, err := io.ReadJSON[T](path)
	if err != nil {
		return nil, err
	}
	cfgMap, err := dict.FromStruct[T, map[string]int](cfg)
	if err != nil {
		return nil, err
	}
	for key := range cfgMap {
		for cfgKey, value := range cfgMap[key] {
			if value == 0 {
				return nil, fmt.Errorf("invalid daemon %s.%s: %d", key, cfgKey, value)
			}
		}
	}
	return cfg, nil
}

// Runs a task every given interval,
// TimeScale = time.Hour, time.Minute, time.Second
func Run(name string, task func(), interval int, timeScale time.Duration) {
	if interval < 0 {
		fmt.Printf("Daemon:%s is disabled\n", name)
		return
	}
	timeInterval := time.Duration(interval) * timeScale
	daemonStart.Set(name, clock.Now())
	daemonDuration.Set(name, timeInterval)
	go func() {
		for {
			start := clock.Now()
			daemonLast.Set(name, start)
			task()
			clock.Sleep(timeInterval, start)
		}
	}()
}

type Info struct {
	Start    string
	Last     string
	Duration string
}

// Returns info on all running daemons
func All() map[string]Info {
	startTime := daemonStart.Map()
	lastTime := daemonLast.Map()
	durations := daemonDuration.Map()
	info := make(map[string]Info)
	for name, startTime := range startTime {
		start := clock.StandardFormat(startTime)
		var last, duration string
		if dict.HasKey(lastTime, name) {
			last = clock.StandardFormat(lastTime[name])
		}
		if dict.HasKey(durations, name) {
			duration = fmt.Sprintf("%v", durations[name])
		}
		info[name] = Info{start, last, duration}
	}
	return info
}
