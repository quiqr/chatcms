package runner

import (
	"fmt"
	"strings"
)

const DefaultLogBytes = 1 << 20

type tailWriter struct {
	max     int
	kept    []byte
	dropped int
}

func (w *tailWriter) Write(p []byte) (int, error) {
	written := len(p)
	if w.max <= 0 {
		w.kept = append(w.kept, p...)
		return written, nil
	}
	if len(p) > w.max {
		w.dropped += len(w.kept) + len(p) - w.max
		w.kept = append(w.kept[:0], p[len(p)-w.max:]...)
		return written, nil
	}
	w.kept = append(w.kept, p...)
	if over := len(w.kept) - w.max; over > 0 {
		w.dropped += over
		w.kept = append(w.kept[:0], w.kept[over:]...)
	}
	return written, nil
}

func (w *tailWriter) String() string {
	if w.dropped == 0 {
		return string(w.kept)
	}
	return Shortened(w.dropped) + string(w.kept)
}

func Shortened(dropped int) string {
	return fmt.Sprintf("[%d earlier bytes of output were dropped; the end is kept]\n", dropped)
}

func TrimLog(log string, max int) string {
	if max <= 0 || len(log) <= max {
		return log
	}
	tail := log[len(log)-max:]
	if cut := strings.IndexByte(tail, '\n'); cut >= 0 && cut < len(tail)-1 {
		tail = tail[cut+1:]
	}
	return Shortened(len(log)-len(tail)) + tail
}
