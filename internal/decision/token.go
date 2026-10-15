package decision

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	// tokenResource is the audience Foundry accepts for Entra ID tokens.
	tokenResource = "https://cognitiveservices.azure.com"
	// refreshMargin is how long before expiry a cached token is replaced.
	refreshMargin = 5 * time.Minute
)

// tokenSource fetches Entra ID access tokens through the Azure CLI, so whoever
// has run `az login` and holds access to the Foundry resource needs no key.
// Tokens are cached on disk because the Read hook starts a fresh process per
// call, and an `az` round trip costs one to two seconds.
type tokenSource struct {
	mu    sync.Mutex
	token string
	exp   time.Time
}

type cachedToken struct {
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}

func cachePath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "systemone", "entra-token.json")
}

func (t *tokenSource) valid() bool {
	return t.token != "" && time.Until(t.exp) > refreshMargin
}

// Token returns a bearer token, refreshing it when it is missing or close to
// expiry.
func (t *tokenSource) Token(ctx context.Context) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.valid() {
		return t.token, nil
	}
	if p := cachePath(); p != "" {
		if raw, err := os.ReadFile(p); err == nil {
			var c cachedToken
			if json.Unmarshal(raw, &c) == nil {
				t.token, t.exp = c.Token, time.Unix(c.Expires, 0)
				if t.valid() {
					return t.token, nil
				}
			}
		}
	}
	return t.refresh(ctx)
}

func (t *tokenSource) refresh(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "az", "account", "get-access-token",
		"--resource", tokenResource, "-o", "json").Output()
	if err != nil {
		hint := "is the Azure CLI installed and on PATH?"
		if ee, ok := err.(*exec.ExitError); ok {
			hint = strings.TrimSpace(string(ee.Stderr))
		}
		return "", fmt.Errorf("getting an Entra ID token failed (run `az login`, or set SYSTEMONE_AUTH=key with AZURE_API_KEY): %s", truncate(hint, 300))
	}
	var resp struct {
		AccessToken string `json:"accessToken"`
		ExpiresOn   int64  `json:"expires_on"`
	}
	if err := json.Unmarshal(out, &resp); err != nil || resp.AccessToken == "" {
		return "", fmt.Errorf("unexpected output from az account get-access-token")
	}
	exp := time.Unix(resp.ExpiresOn, 0)
	if resp.ExpiresOn == 0 {
		exp = time.Now().Add(30 * time.Minute) // older az CLIs omit expires_on
	}
	t.token, t.exp = resp.AccessToken, exp

	if p := cachePath(); p != "" {
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err == nil {
			if raw, err := json.Marshal(cachedToken{Token: t.token, Expires: exp.Unix()}); err == nil {
				_ = os.WriteFile(p, raw, 0o600)
			}
		}
	}
	return t.token, nil
}

// Invalidate drops the cached token, in memory and on disk, after a 401.
func (t *tokenSource) Invalidate() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.token, t.exp = "", time.Time{}
	if p := cachePath(); p != "" {
		_ = os.Remove(p)
	}
}
