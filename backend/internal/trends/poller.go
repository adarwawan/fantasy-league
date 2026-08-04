package trends

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Poller captures snapshots on a fixed interval while a session is active. It
// dedups against the last-seen value per player (in memory, reset per gameweek)
// so unchanged players don't fabricate motion or waste rows.
type Poller struct {
	client *Client
	store  *Store

	mu       sync.Mutex
	lastGW   int
	lastSeen map[int]Snapshot // player_ext_id -> last stored snapshot for lastGW
}

func NewPoller(client *Client, store *Store) *Poller {
	return &Poller{client: client, store: store, lastSeen: map[int]Snapshot{}}
}

// Tick runs one poll. It no-ops when no session is active. Safe to call from a
// scheduler; errors are logged, not returned, so the schedule keeps running.
func (p *Poller) Tick(ctx context.Context) {
	sess, err := p.store.ActiveSession(ctx)
	if err != nil {
		slog.Error("trends: active session check failed", "err", err)
		return
	}
	if sess == nil {
		return
	}

	snaps, err := p.client.FetchSnapshots(ctx)
	if err != nil {
		slog.Error("trends: fetch snapshots failed", "err", err)
		return
	}

	changed := p.dedup(sess.Gameweek, snaps)
	capturedAt := time.Now().UTC()
	if err := p.store.InsertSnapshots(ctx, sess.Gameweek, capturedAt, changed); err != nil {
		slog.Error("trends: insert snapshots failed", "err", err)
		return
	}
	slog.Info("trends: poll captured", "gw", sess.Gameweek, "changed", len(changed), "total", len(snaps))
}

// dedup returns only the snapshots whose captured fields differ from the last
// stored value. The per-player baseline resets when the gameweek changes.
func (p *Poller) dedup(gw int, snaps []Snapshot) []Snapshot {
	p.mu.Lock()
	defer p.mu.Unlock()
	if gw != p.lastGW {
		p.lastGW = gw
		p.lastSeen = make(map[int]Snapshot, len(snaps))
	}
	changed := make([]Snapshot, 0, len(snaps))
	for _, sn := range snaps {
		if prev, ok := p.lastSeen[sn.PlayerExtID]; ok && prev == sn {
			continue
		}
		p.lastSeen[sn.PlayerExtID] = sn
		changed = append(changed, sn)
	}
	return changed
}
