package trends

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const fplBase = "https://fantasy.premierleague.com/api"

// Client fetches the ownership/transfer fields from bootstrap-static. It is a
// tiny self-contained client (like setpiece's) so the poller's dependency on
// FPL degrades independently of the main app.
type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient() *Client {
	return &Client{baseURL: fplBase, http: &http.Client{Timeout: 15 * time.Second}}
}

type bootstrapResponse struct {
	Elements []element `json:"elements"`
	Events   []gwEvent `json:"events"`
}

type element struct {
	ID                int    `json:"id"`
	SelectedByPercent string `json:"selected_by_percent"` // "12.3"
	TransfersInEvent  int    `json:"transfers_in_event"`
	TransfersOutEvent int    `json:"transfers_out_event"`
	NowCost           int    `json:"now_cost"`
}

type gwEvent struct {
	ID           int    `json:"id"`
	IsCurrent    bool   `json:"is_current"`
	IsNext       bool   `json:"is_next"`
	DeadlineTime string `json:"deadline_time"`
}

// FetchSnapshots returns one Snapshot per player from a single bootstrap call.
func (c *Client) FetchSnapshots(ctx context.Context) ([]Snapshot, error) {
	b, err := c.fetchBootstrap(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Snapshot, 0, len(b.Elements))
	for _, e := range b.Elements {
		pct, _ := strconv.ParseFloat(e.SelectedByPercent, 64)
		out = append(out, Snapshot{
			PlayerExtID:  e.ID,
			SelectedBP:   int(pct*100 + 0.5),
			TransfersIn:  e.TransfersInEvent,
			TransfersOut: e.TransfersOutEvent,
			NowCost:      e.NowCost,
		})
	}
	return out, nil
}

// NextDeadline returns the (gameweek, deadline) of the upcoming FPL event, used
// to auto-fill session arming. Falls back to the current event if none is next.
func (c *Client) NextDeadline(ctx context.Context) (gw int, deadline time.Time, err error) {
	b, err := c.fetchBootstrap(ctx)
	if err != nil {
		return 0, time.Time{}, err
	}
	for _, e := range b.Events {
		if e.IsNext {
			t, _ := time.Parse(time.RFC3339, e.DeadlineTime)
			return e.ID, t, nil
		}
	}
	for _, e := range b.Events {
		if e.IsCurrent {
			t, _ := time.Parse(time.RFC3339, e.DeadlineTime)
			return e.ID, t, nil
		}
	}
	return 0, time.Time{}, fmt.Errorf("no current or next event in bootstrap")
}

func (c *Client) fetchBootstrap(ctx context.Context) (*bootstrapResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/bootstrap-static/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "fantasy-league-dashboard/0.1")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET bootstrap-static: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET bootstrap-static: status %d", resp.StatusCode)
	}
	var r bootstrapResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decode bootstrap: %w", err)
	}
	return &r, nil
}
