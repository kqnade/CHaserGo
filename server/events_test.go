package server

import (
	"testing"
)

func TestSnapshotFromBoard(t *testing.T) {
	b := newTestBoard()
	b.Hot.Name = "PlayerHot"
	b.Cool.Name = "PlayerCool"
	b.Hot.Items = 3
	b.Cool.Items = 1
	b.Turn = 5

	snap := SnapshotFromBoard(b, KindActionEnd, TurnStepFirst, PhaseRunning, 42, "PlayerHot", "hot wins")

	t.Run("メタデータ", func(t *testing.T) {
		if snap.Kind != KindActionEnd {
			t.Errorf("Kind = %v, want KindActionEnd", snap.Kind)
		}
		if snap.Step != TurnStepFirst {
			t.Errorf("Step = %v, want TurnStepFirst", snap.Step)
		}
		if snap.Phase != PhaseRunning {
			t.Errorf("Phase = %v, want PhaseRunning", snap.Phase)
		}
		if snap.Revision != 42 {
			t.Errorf("Revision = %d, want 42", snap.Revision)
		}
		if snap.WinnerName != "PlayerHot" {
			t.Errorf("WinnerName = %q, want \"PlayerHot\"", snap.WinnerName)
		}
		if snap.Reason != "hot wins" {
			t.Errorf("Reason = %q, want \"hot wins\"", snap.Reason)
		}
	})

	t.Run("盤面サイズとターン", func(t *testing.T) {
		if snap.Width != 5 || snap.Height != 5 {
			t.Errorf("size = %dx%d, want 5x5", snap.Width, snap.Height)
		}
		if snap.MaxTurns != 100 {
			t.Errorf("MaxTurns = %d, want 100", snap.MaxTurns)
		}
		if snap.Turn != 5 {
			t.Errorf("Turn = %d, want 5", snap.Turn)
		}
	})

	t.Run("Hot キャラクター情報", func(t *testing.T) {
		if snap.HotName != "PlayerHot" {
			t.Errorf("HotName = %q, want \"PlayerHot\"", snap.HotName)
		}
		if snap.HotX != 1 || snap.HotY != 1 {
			t.Errorf("Hot pos = (%d,%d), want (1,1)", snap.HotX, snap.HotY)
		}
		if snap.HotItems != 3 {
			t.Errorf("HotItems = %d, want 3", snap.HotItems)
		}
		if !snap.HotAlive {
			t.Error("HotAlive should be true")
		}
	})

	t.Run("Cool キャラクター情報", func(t *testing.T) {
		if snap.CoolName != "PlayerCool" {
			t.Errorf("CoolName = %q, want \"PlayerCool\"", snap.CoolName)
		}
		if snap.CoolX != 3 || snap.CoolY != 3 {
			t.Errorf("Cool pos = (%d,%d), want (3,3)", snap.CoolX, snap.CoolY)
		}
		if snap.CoolItems != 1 {
			t.Errorf("CoolItems = %d, want 1", snap.CoolItems)
		}
		if !snap.CoolAlive {
			t.Error("CoolAlive should be true")
		}
	})

	t.Run("MapFlat の内容確認", func(t *testing.T) {
		// (y=0, x=0) = Wall = 2
		if snap.MapFlat[0] != int(Wall) {
			t.Errorf("MapFlat[0] = %d, want %d (Wall)", snap.MapFlat[0], int(Wall))
		}
		// (y=2, x=2) = Item = 3
		if snap.MapFlat[2*5+2] != int(Item) {
			t.Errorf("MapFlat[2*5+2] = %d, want %d (Item)", snap.MapFlat[2*5+2], int(Item))
		}
	})

	t.Run("MapFlat は深いコピー", func(t *testing.T) {
		if len(snap.MapFlat) != b.Width*b.Height {
			t.Errorf("MapFlat len = %d, want %d", len(snap.MapFlat), b.Width*b.Height)
		}
		// 元の Board を変更してもスナップショットに影響しないこと
		original := snap.MapFlat[2*5+2] // (y=2, x=2) = Item
		b.MapData[2][2] = Wall
		if snap.MapFlat[2*5+2] != original {
			t.Error("MapFlat should be a deep copy")
		}
	})
}

func TestSnapshotKindConstants(t *testing.T) {
	// 定数の順序が変わっていないことを確認
	kinds := []SnapshotKind{KindInitial, KindConnected, KindActionEnd, KindTurnEnd, KindGameOver, KindError}
	for i, k := range kinds {
		if int(k) != i {
			t.Errorf("SnapshotKind[%d] = %d, want %d", i, k, i)
		}
	}
}

func TestTurnStepConstants(t *testing.T) {
	if TurnStepFirst != 0 {
		t.Errorf("TurnStepFirst = %d, want 0", TurnStepFirst)
	}
	if TurnStepSecond != 1 {
		t.Errorf("TurnStepSecond = %d, want 1", TurnStepSecond)
	}
}

func TestSnapshotPublicPhaseConstants(t *testing.T) {
	phases := []SnapshotPublicPhase{PhaseWaiting, PhaseRunning, PhaseGameOver, PhaseError}
	for i, p := range phases {
		if int(p) != i {
			t.Errorf("SnapshotPublicPhase[%d] = %d, want %d", i, p, i)
		}
	}
}
