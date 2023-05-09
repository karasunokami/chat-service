package schema

import "time"

func defaultTime() time.Time {
	return time.Now().Truncate(time.Millisecond)
}
