package controllers

import "testing"

func TestAssetPaginationForTotal(t *testing.T) {
	meta := assetPaginationForTotal(2, 50, 125)
	if meta.PageStart != 51 || meta.PageEnd != 100 || meta.TotalPages != 3 || !meta.HasPrev || !meta.HasNext {
		t.Fatalf("unexpected page metadata: %+v", meta)
	}

	last := assetPaginationForTotal(99, 50, 125)
	if last.CurrentPage != 3 || last.PageStart != 101 || last.PageEnd != 125 || last.HasNext {
		t.Fatalf("unexpected clamped metadata: %+v", last)
	}

	empty := assetPaginationForTotal(1, 50, 0)
	if empty.PageStart != 0 || empty.PageEnd != 0 || empty.TotalPages != 1 {
		t.Fatalf("unexpected empty metadata: %+v", empty)
	}
}
