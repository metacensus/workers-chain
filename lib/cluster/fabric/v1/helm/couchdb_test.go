package helm

import "testing"

// TestCouchDBChart_CapsPortTable renders the CouchDB chart as the worker does:
// the couchdb container must cap Erlang's port table. couchdb:3.3.2's Erlang
// sizes that table from the open-file limit, and on a runtime that sets it to
// 1073741816 (Docker Desktop's kind) allocates ~1.5GB at boot and is
// OOMKilled (#54). vm.args sets no +Q, so ERL_FLAGS is what applies.
func TestCouchDBChart_CapsPortTable(t *testing.T) {
	couchdb := renderContainer(t, "fabric/couchdb", "couchdb")
	for _, env := range couchdb.Env {
		if env.Name == "ERL_FLAGS" {
			if env.Value != "+Q 65536" {
				t.Errorf("ERL_FLAGS = %q, want %q", env.Value, "+Q 65536")
			}
			return
		}
	}
	t.Error("couchdb container does not set ERL_FLAGS")
}
