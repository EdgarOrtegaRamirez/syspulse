package alerts

import "testing"

func TestConfig_Defaults(t *testing.T) {
	cfg := &Config{}
	if cfg.CPU != 0 {
		t.Errorf("expected default CPU threshold 0, got %d", cfg.CPU)
	}
}
