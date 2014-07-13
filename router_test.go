package bayeux

import (
	"testing"
)

func TestNewRouter(t *testing.T) {
	router := NewRouter()

	if router == nil {
		t.Error("Router Was Not Created")
	}
}
