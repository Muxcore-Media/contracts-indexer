package indexerv1

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestIndexerServiceRPCNamesStable(t *testing.T) {
	want := map[string]bool{
		"Search":          true,
		"GetCapabilities": true,
		"ListIndexers":    true,
		"GetIndexer":      true,
		"CreateIndexer":   true,
		"UpdateIndexer":   true,
		"DeleteIndexer":   true,
	}
	got := make(map[string]bool, len(IndexerService_ServiceDesc.Methods))
	for _, m := range IndexerService_ServiceDesc.Methods {
		got[m.MethodName] = true
	}
	for name := range want {
		if !got[name] {
			t.Errorf("missing RPC %s", name)
		}
	}
}

func TestSearchRequestFieldNumbers(t *testing.T) {
	want := map[string]protoreflect.FieldNumber{
		"query":       1,
		"type":        2,
		"year":        3,
		"season":      4,
		"episode":     5,
		"categories":  6,
		"limit":       7,
		"offset":      8,
		"indexer_ids": 9,
		"absolute":    10,
		"imdb_id":     11,
		"tmdb_id":     12,
		"tvdb_id":     13,
	}
	assertFieldNumbers(t, (&SearchRequest{}).ProtoReflect().Descriptor(), want)
}

func TestSearchResultFieldNumbers(t *testing.T) {
	want := map[string]protoreflect.FieldNumber{
		"guid":              1,
		"title":             2,
		"info_url":          3,
		"download_url":      4,
		"size":              5,
		"seeders":           6,
		"peers":             7,
		"indexer_name":      8,
		"indexer_id":        9,
		"publish_date":      10,
		"category":          11,
		"sub_category":      12,
		"imdb_id":           13,
		"tmdb_id":           14,
		"tvdb_id":           15,
		"download_protocol": 16,
		"infohash":          17,
	}
	assertFieldNumbers(t, (&SearchResult{}).ProtoReflect().Descriptor(), want)
}

func TestIndexerInfoFieldNumbers(t *testing.T) {
	want := map[string]protoreflect.FieldNumber{
		"id":         1,
		"name":       2,
		"protocol":   3,
		"language":   4,
		"configured": 5,
	}
	assertFieldNumbers(t, (&IndexerInfo{}).ProtoReflect().Descriptor(), want)
}

func TestIndexerSpecFieldNumbers(t *testing.T) {
	want := map[string]protoreflect.FieldNumber{
		"id":             1,
		"name":           2,
		"protocol":       3,
		"implementation": 4,
		"base_url":       5,
		"api_key":        6,
		"enable":         7,
		"has_api_key":    8,
		"language":       9,
	}
	assertFieldNumbers(t, (&IndexerSpec{}).ProtoReflect().Descriptor(), want)
}

func TestTVSeasonPackAndAnimeAbsoluteFieldTags(t *testing.T) {
	desc := (&SearchRequest{}).ProtoReflect().Descriptor()
	if got := desc.Fields().ByName("season").Number(); got != 4 {
		t.Fatalf("season field tag = %d, want 4", got)
	}
	if got := desc.Fields().ByName("episode").Number(); got != 5 {
		t.Fatalf("episode field tag = %d, want 5", got)
	}
	if got := desc.Fields().ByName("absolute").Number(); got != 10 {
		t.Fatalf("absolute field tag = %d, want 10", got)
	}
}

func TestSearchTypeVocabulary(t *testing.T) {
	cases := []SearchType{
		SearchType_SEARCH_TYPE_MOVIE,
		SearchType_SEARCH_TYPE_TV,
		SearchType_SEARCH_TYPE_MUSIC,
		SearchType_SEARCH_TYPE_BOOK,
		SearchType_SEARCH_TYPE_AUDIOBOOK,
		SearchType_SEARCH_TYPE_COMIC,
	}
	for _, st := range cases {
		if st == SearchType_SEARCH_TYPE_UNSPECIFIED {
			t.Errorf("unexpected unspecified in vocabulary list")
		}
		if st.String() == "" {
			t.Errorf("search type %v has empty String()", st)
		}
	}
}

func TestImportPath(t *testing.T) {
	var _ = IndexerService_ServiceDesc
}

func TestCapabilityIDs(t *testing.T) {
	const consumerCapability = "indexer"
	const contractCapability = "contracts.indexer"
	if consumerCapability == contractCapability {
		t.Fatal("consumer and contract capability ids must differ")
	}
}

func assertFieldNumbers(t *testing.T, desc protoreflect.MessageDescriptor, want map[string]protoreflect.FieldNumber) {
	t.Helper()
	for name, num := range want {
		fd := desc.Fields().ByName(protoreflect.Name(name))
		if fd == nil {
			t.Errorf("missing field %q", name)
			continue
		}
		if fd.Number() != num {
			t.Errorf("field %q number = %d, want %d", name, fd.Number(), num)
		}
	}
}
