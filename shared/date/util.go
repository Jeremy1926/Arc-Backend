package date

import "time"

func UTC() time.Time {
	return time.Now().UTC()
}

func RFC3339() string {
	return UTC().Format(time.RFC3339)
}
