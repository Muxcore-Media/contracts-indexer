// Package conformance is a reusable test suite that every IndexerService
// implementation should pass. It checks the rules documented in indexer.proto
// that the generated types cannot enforce: result limits and totals, no
// credentials in download URLs, and how CRUD behaves on modules without it.
//
// Usage from a module's tests, against an in-process server or fixture feed:
//
//	func TestConformance(t *testing.T) {
//		conformance.Run(t, conformance.Suite{
//			Client: newTestClient(t),
//			Query:  "Big Buck Bunny",
//		})
//	}
package conformance

import (
	"context"
	"strings"
	"testing"
	"time"

	indexerv1 "github.com/Muxcore-Media/contracts-indexer/muxcore/indexer/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultLimit = 50
	maxLimit     = 100
)

// Suite configures a conformance run.
type Suite struct {
	// Client talks to the implementation under test.
	Client indexerv1.IndexerServiceClient

	// Query is a search the implementation's fixture data answers with at
	// least one result. Empty skips the search checks.
	Query string
	// Type is the SearchRequest.type for Query (movie, tv, ...); optional.
	Type string

	// SupportsCRUD is true for implementations that manage indexers (Prowlarr
	// mode). Others must answer the CRUD RPCs with FailedPrecondition or
	// Unimplemented.
	SupportsCRUD bool
	// SampleIndexer is created, read back and deleted when SupportsCRUD is set.
	// Give it an api_key so the write-only rule is checked.
	SampleIndexer *indexerv1.IndexerSpec

	// Timeout bounds each RPC. Defaults to 30s (searches hit real feeds).
	Timeout time.Duration
}

// Run executes the suite as subtests of t.
func Run(t *testing.T, s Suite) {
	t.Helper()
	if s.Client == nil {
		t.Fatal("conformance: Suite.Client is required")
	}
	if s.Timeout <= 0 {
		s.Timeout = 30 * time.Second
	}

	var caps *indexerv1.GetCapabilitiesResponse
	t.Run("GetCapabilities", func(t *testing.T) {
		ctx, cancel := s.ctx()
		defer cancel()
		resp, err := s.Client.GetCapabilities(ctx, &indexerv1.GetCapabilitiesRequest{})
		if err != nil {
			t.Fatalf("GetCapabilities: %v", err)
		}
		for _, p := range resp.GetSupportedProtocols() {
			if p != "torrent" && p != "usenet" {
				t.Errorf("supported_protocols has %q; want torrent or usenet", p)
			}
		}
		caps = resp
	})

	t.Run("ListIndexers", func(t *testing.T) {
		ctx, cancel := s.ctx()
		defer cancel()
		if _, err := s.Client.ListIndexers(ctx, &indexerv1.ListIndexersRequest{}); err != nil {
			t.Fatalf("ListIndexers: %v", err)
		}
	})

	if s.Query != "" && (caps == nil || caps.GetSupportsSearch()) {
		t.Run("Search", s.search)
		t.Run("SearchHonorsLimit", s.searchLimit)
	} else {
		t.Log("conformance: no query or search unsupported; skipping search checks")
	}

	if s.SupportsCRUD {
		t.Run("CRUD", s.crud)
	} else {
		t.Run("CRUDUnsupported", s.crudUnsupported)
	}
}

func (s Suite) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.Timeout)
}

func (s Suite) search(t *testing.T) {
	ctx, cancel := s.ctx()
	defer cancel()
	resp, err := s.Client.Search(ctx, &indexerv1.SearchRequest{Query: s.Query, Type: s.Type})
	if err != nil {
		t.Fatalf("Search(%q): %v", s.Query, err)
	}
	results := resp.GetResults()
	if len(results) == 0 {
		t.Fatalf("Search(%q) returned no results; the suite needs fixture data", s.Query)
	}
	if len(results) > defaultLimit {
		t.Errorf("unset limit returned %d results; default is %d", len(results), defaultLimit)
	}
	if int(resp.GetTotal()) < len(results) {
		t.Errorf("total=%d is less than len(results)=%d; total is the unpaged hit count", resp.GetTotal(), len(results))
	}
	for i, r := range results {
		checkResult(t, i, r)
	}
}

func (s Suite) searchLimit(t *testing.T) {
	ctx, cancel := s.ctx()
	defer cancel()
	resp, err := s.Client.Search(ctx, &indexerv1.SearchRequest{Query: s.Query, Type: s.Type, Limit: 1})
	if err != nil {
		t.Fatalf("Search(limit=1): %v", err)
	}
	if n := len(resp.GetResults()); n > 1 {
		t.Errorf("limit=1 returned %d results", n)
	}
	resp, err = s.Client.Search(ctx, &indexerv1.SearchRequest{Query: s.Query, Type: s.Type, Limit: 1000})
	if err != nil {
		t.Fatalf("Search(limit=1000): %v", err)
	}
	if n := len(resp.GetResults()); n > maxLimit {
		t.Errorf("limit=1000 returned %d results; max is %d", n, maxLimit)
	}
}

func checkResult(t *testing.T, i int, r *indexerv1.SearchResult) {
	t.Helper()
	if strings.TrimSpace(r.GetTitle()) == "" {
		t.Errorf("result %d has no title", i)
	}
	if r.GetGuid() == "" && r.GetDownloadUrl() == "" {
		t.Errorf("result %d (%q) has neither guid nor download_url", i, r.GetTitle())
	}
	if leaksCredentials(r.GetDownloadUrl()) {
		t.Errorf("result %d download_url embeds credentials: %q", i, r.GetDownloadUrl())
	}
	if p := r.GetDownloadProtocol(); p != "" && p != "torrent" && p != "usenet" {
		t.Errorf("result %d download_protocol=%q; want torrent, usenet or empty", i, p)
	}
	if r.GetSize() < 0 || r.GetSeeders() < 0 || r.GetPeers() < 0 {
		t.Errorf("result %d has negative size/seeders/peers", i)
	}
}

// leaksCredentials flags Torznab/Newznab API keys and URL userinfo.
func leaksCredentials(u string) bool {
	l := strings.ToLower(u)
	for _, key := range []string{"apikey=", "api_key=", "passkey="} {
		if strings.Contains(l, key) {
			return true
		}
	}
	if i := strings.Index(l, "://"); i >= 0 {
		host := l[i+3:]
		if j := strings.IndexAny(host, "/?#"); j >= 0 {
			host = host[:j]
		}
		return strings.Contains(host, "@")
	}
	return false
}

func (s Suite) crudUnsupported(t *testing.T) {
	ctx, cancel := s.ctx()
	defer cancel()
	_, err := s.Client.GetIndexer(ctx, &indexerv1.GetIndexerRequest{Id: 1})
	switch status.Code(err) {
	case codes.FailedPrecondition, codes.Unimplemented:
	default:
		t.Errorf("GetIndexer without CRUD support: %v; want FailedPrecondition or Unimplemented", err)
	}
}

func (s Suite) crud(t *testing.T) {
	if s.SampleIndexer == nil {
		t.Fatal("conformance: SupportsCRUD needs SampleIndexer")
	}
	ctx, cancel := s.ctx()
	defer cancel()
	created, err := s.Client.CreateIndexer(ctx, &indexerv1.CreateIndexerRequest{Indexer: s.SampleIndexer})
	if err != nil {
		t.Fatalf("CreateIndexer: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := s.ctx()
		defer cancel()
		_, _ = s.Client.DeleteIndexer(ctx, &indexerv1.DeleteIndexerRequest{Id: created.GetId()})
	})
	checkSpec(t, "CreateIndexer", s.SampleIndexer, created)

	got, err := s.Client.GetIndexer(ctx, &indexerv1.GetIndexerRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetIndexer(%d): %v", created.GetId(), err)
	}
	checkSpec(t, "GetIndexer", s.SampleIndexer, got)

	if _, err := s.Client.DeleteIndexer(ctx, &indexerv1.DeleteIndexerRequest{Id: created.GetId()}); err != nil {
		t.Fatalf("DeleteIndexer: %v", err)
	}
	_, err = s.Client.GetIndexer(ctx, &indexerv1.GetIndexerRequest{Id: created.GetId()})
	if err == nil {
		t.Error("GetIndexer after delete must fail")
	} else if status.Code(err) != codes.NotFound {
		t.Logf("GetIndexer after delete: prefer gRPC NotFound (got %v)", status.Code(err))
	}
}

func checkSpec(t *testing.T, call string, sent, got *indexerv1.IndexerSpec) {
	t.Helper()
	if got.GetId() == 0 {
		t.Errorf("%s returned id 0", call)
	}
	if got.GetApiKey() != "" {
		t.Errorf("%s echoed api_key; it is write-only", call)
	}
	if sent.GetApiKey() != "" && !got.GetHasApiKey() {
		t.Errorf("%s: has_api_key=false after an api_key was set", call)
	}
	if got.GetName() != sent.GetName() {
		t.Errorf("%s name=%q, want %q", call, got.GetName(), sent.GetName())
	}
}
