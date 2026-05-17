package statsapi

import (
	"testing"
)

func TestPlayerFromCacheReturnsStoredEntry(t *testing.T) {
	client := &MLBClient{
		playerLookup: map[int32]BaseballPersonRestObject{
			123: {Id: int32Ptr(123)},
		},
	}

	player, ok := client.playerFromLookup(123)
	if !ok {
		t.Fatal("expected cached player hit")
	}
	if player.Id == nil || *player.Id != 123 {
		t.Fatalf("expected cached player id 123, got %+v", player)
	}
}

func TestBuildPlayerLookupSkipsNilIDs(t *testing.T) {
	lookup := buildPlayerLookup([]BaseballPersonRestObject{
		{Id: int32Ptr(123)},
		{},
		{Id: int32Ptr(456)},
	})

	if len(lookup) != 2 {
		t.Fatalf("expected 2 players in lookup, got %d", len(lookup))
	}
	if _, ok := lookup[123]; !ok {
		t.Fatal("expected player 123 in lookup")
	}
	if _, ok := lookup[456]; !ok {
		t.Fatal("expected player 456 in lookup")
	}
}

func TestPlayerLookupSnapshotReturnsCopy(t *testing.T) {
	client := &MLBClient{
		playerLookup: map[int32]BaseballPersonRestObject{
			123: {Id: int32Ptr(123)},
		},
	}

	snapshot := client.PlayerLookupSnapshot()
	delete(snapshot, 123)

	if _, ok := client.playerLookup[123]; !ok {
		t.Fatal("expected snapshot mutation not to affect player lookup")
	}
}

func TestStorePlayerInLookup(t *testing.T) {
	client := &MLBClient{playerLookup: make(map[int32]BaseballPersonRestObject)}
	player := BaseballPersonRestObject{Id: int32Ptr(321)}

	client.storePlayerInLookup(321, player)

	got, ok := client.playerLookup[321]
	if !ok {
		t.Fatal("expected player to be stored in lookup")
	}
	if got.Id == nil || *got.Id != 321 {
		t.Fatalf("expected player id 321, got %+v", got)
	}
}

func int32Ptr(v int32) *int32 { return &v }
