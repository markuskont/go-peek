package trackers

import (
	"fmt"
	"time"
)

type AssetFirstSeenOnWire struct {
	Timestamp time.Time
	Name      string
}

func (a AssetFirstSeenOnWire) String() string {
	return fmt.Sprintf("[%s] asset %s first seen", a.Timestamp.String(), a.Name)
}
