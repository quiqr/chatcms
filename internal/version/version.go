package version

import (
	"fmt"
	"runtime"
)

var Version = "dev"

func String() string {
	return fmt.Sprintf("chatcmsd %s (%s/%s, %s)",
		Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
}
