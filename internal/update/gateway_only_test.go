package update

import (
	"strings"
	"testing"
)

func TestGatewayOnlyUpdatesStayOnTheFork(t *testing.T) {
	t.Setenv("MAGPIE_UPDATE_FEED", "")
	t.Setenv("MAGPIE_NOTES_FEED", "")
	for _, u := range []string{Feed(), NotesFeed(), ReleasePage("0.1.0-gateway.1")} {
		if !strings.HasPrefix(u, "https://github.com/AlexbeatsZ/magpie/releases/") {
			t.Errorf("fork could reinstall the removed feature through %s", u)
		}
	}
	t.Setenv("MAGPIE_UPDATE_FEED", "http://127.0.0.1/test-feed")
	if Feed() != "http://127.0.0.1/test-feed" {
		t.Fatal("explicit test/deployment feed override lost")
	}
}
