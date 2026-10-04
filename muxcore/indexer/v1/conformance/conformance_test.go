package conformance

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"

	indexerv1 "github.com/Muxcore-Media/contracts-indexer/muxcore/indexer/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// memIndexer is a minimal reference implementation the suite must accept.
type memIndexer struct {
	indexerv1.UnimplementedIndexerServiceServer
	crud bool

	mu     sync.Mutex
	nextID int32
	specs  map[int32]*indexerv1.IndexerSpec
}

func (m *memIndexer) GetCapabilities(context.Context, *indexerv1.GetCapabilitiesRequest) (*indexerv1.GetCapabilitiesResponse, error) {
	return &indexerv1.GetCapabilitiesResponse{SupportsSearch: true, SupportedProtocols: []string{"torrent"}}, nil
}

func (m *memIndexer) ListIndexers(context.Context, *indexerv1.ListIndexersRequest) (*indexerv1.ListIndexersResponse, error) {
	return &indexerv1.ListIndexersResponse{Indexers: []*indexerv1.IndexerInfo{{Id: 1, Name: "fixture", Protocol: "torrent", Configured: true}}}, nil
}

func (m *memIndexer) Search(_ context.Context, in *indexerv1.SearchRequest) (*indexerv1.SearchResponse, error) {
	const hits = 120
	limit := int(in.GetLimit())
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	out := &indexerv1.SearchResponse{Total: hits}
	for i := 0; i < hits && i < limit; i++ {
		out.Results = append(out.Results, &indexerv1.SearchResult{
			Guid:             fmt.Sprintf("guid-%d", i),
			Title:            fmt.Sprintf("%s.%d.1080p", in.GetQuery(), i),
			DownloadUrl:      fmt.Sprintf("https://tracker.example/dl/%d", i),
			DownloadProtocol: "torrent",
			Seeders:          10,
		})
	}
	return out, nil
}

func (m *memIndexer) GetIndexer(_ context.Context, in *indexerv1.GetIndexerRequest) (*indexerv1.IndexerSpec, error) {
	if !m.crud {
		return nil, status.Error(codes.FailedPrecondition, "indexer CRUD requires Prowlarr")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	spec, ok := m.specs[in.GetId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "indexer %d not found", in.GetId())
	}
	return redact(spec), nil
}

func (m *memIndexer) CreateIndexer(_ context.Context, in *indexerv1.CreateIndexerRequest) (*indexerv1.IndexerSpec, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	spec := &indexerv1.IndexerSpec{
		Id:     m.nextID,
		Name:   in.GetIndexer().GetName(),
		ApiKey: in.GetIndexer().GetApiKey(),
	}
	m.specs[spec.Id] = spec
	return redact(spec), nil
}

func (m *memIndexer) DeleteIndexer(_ context.Context, in *indexerv1.DeleteIndexerRequest) (*indexerv1.DeleteIndexerResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.specs, in.GetId())
	return &indexerv1.DeleteIndexerResponse{}, nil
}

func redact(s *indexerv1.IndexerSpec) *indexerv1.IndexerSpec {
	return &indexerv1.IndexerSpec{Id: s.GetId(), Name: s.GetName(), HasApiKey: s.GetApiKey() != ""}
}

func newMemClient(t *testing.T, impl *memIndexer) indexerv1.IndexerServiceClient {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	indexerv1.RegisterIndexerServiceServer(srv, impl)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return indexerv1.NewIndexerServiceClient(conn)
}

func TestReferenceImplementationPasses(t *testing.T) {
	Run(t, Suite{Client: newMemClient(t, &memIndexer{}), Query: "Big.Buck.Bunny"})
}

func TestReferenceImplementationWithCRUD(t *testing.T) {
	Run(t, Suite{
		Client:        newMemClient(t, &memIndexer{crud: true, specs: map[int32]*indexerv1.IndexerSpec{}}),
		Query:         "Big.Buck.Bunny",
		SupportsCRUD:  true,
		SampleIndexer: &indexerv1.IndexerSpec{Name: "conformance", ApiKey: "secret"},
	})
}

func TestLeaksCredentials(t *testing.T) {
	cases := map[string]bool{
		"https://tracker.example/dl/1":                     false,
		"magnet:?xt=urn:btih:abc":                          false,
		"https://tracker.example/api?t=get&apikey=abc":     true,
		"https://tracker.example/dl/1?passkey=abc":         true,
		"https://user:pass@tracker.example/dl/1":           true,
		"https://tracker.example/dl/1?ref=someone@example": false,
	}
	for u, want := range cases {
		if got := leaksCredentials(u); got != want {
			t.Errorf("leaksCredentials(%q)=%v want %v", u, got, want)
		}
	}
}
