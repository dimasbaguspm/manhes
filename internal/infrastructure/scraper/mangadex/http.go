package mangadex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (a *Adapter) get(ctx context.Context, u string, out any) error {
	if err := a.limiter.Wait(ctx); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "manga-engine/1.0")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d for %s", resp.StatusCode, u)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func firstOf(m map[string]string, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != "" {
			return v
		}
	}
	for _, v := range m {
		return v
	}
	return ""
}
