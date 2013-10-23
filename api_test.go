package heroku

import (
	"testing"
)

func TestListPather(t *testing.T) {
	var apps []*App
	path := listPather(&apps).Path()
	if path != "/apps/" {
		t.Fatal("listPather(*[]*App).Path = %q want %q", path, "/apps/")
	}
}
