package gitsync

import "time"

// Save is one write to a site: the files it changes, the files it removes,
// and who made it.
type Save struct {
	Message string
	Author  Identity
	Writes  map[string][]byte
	Deletes []string
}

// Version is one saved state of a site. The identifier is opaque: nothing
// outside this package may take it apart.
type Version struct {
	ID      string    `json:"id"`
	Time    time.Time `json:"time"`
	Author  Identity  `json:"author"`
	Message string    `json:"message"`
	Paths   []string  `json:"paths,omitempty"`
}

// Status answers the three questions the interface asks: is everything saved,
// how much is waiting to be published, and did someone else change the site.
type Status struct {
	UnsavedWork        bool `json:"unsavedWork"`
	Unpublished        int  `json:"unpublished"`
	RemoteAhead        int  `json:"remoteAhead"`
	ExchangeInProgress bool `json:"exchangeInProgress"`
}

// ConflictedPath carries every version of one file that two people changed.
type ConflictedPath struct {
	Path   string `json:"path"`
	Base   []byte `json:"-"`
	Local  []byte `json:"-"`
	Remote []byte `json:"-"`
}

// Conflict is what a sync reports when it cannot finish on its own.
type Conflict struct {
	Paths []ConflictedPath `json:"paths"`
}

func (c *Conflict) pathNames() []string {
	names := make([]string, 0, len(c.Paths))
	for _, path := range c.Paths {
		names = append(names, path.Path)
	}
	return names
}

// SyncResult says what an exchange with the remote did.
type SyncResult struct {
	Arrived  []Version `json:"arrived,omitempty"`
	Replayed int       `json:"replayed"`
	Conflict *Conflict `json:"conflict,omitempty"`
}

// PublishResult says what reached the remote.
type PublishResult struct {
	Published []Version `json:"published,omitempty"`
}
