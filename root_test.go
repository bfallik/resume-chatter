package root

import (
	"sync"
	"testing"

	"github.com/bfallik/resume-chatter/internal/model"
)

func TestHistorySync(t *testing.T) {
	var wg sync.WaitGroup

	h := History{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if len(h.GetChat()) != 0 && len(h.GetChat()) != 2 {
			t.Errorf("test error: expected 0 or 2, got %v", len(h.GetChat()))
		}
	}()
	h.Append(model.Chat{}, model.Chat{})
	wg.Wait()
}
