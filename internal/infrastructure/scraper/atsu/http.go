package atsu

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
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")
	req.Header.Set("Referer", a.baseURL)

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
