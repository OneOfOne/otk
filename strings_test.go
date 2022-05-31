package otk

import "testing"

func TestFindKeywords(t *testing.T) {
	t.Run("crash on invalid query", func(t *testing.T) {
		defer func() {
			if p := recover(); p != nil {
				t.Errorf("panic: %v", p)
			}
		}()
		KeywordsToSearch(`"a": x:`)
	})
	t.Log()
}
