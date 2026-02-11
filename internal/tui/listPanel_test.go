package tui

import "testing"

func TestCalcCursorDown(t *testing.T) {
	tests := []struct {
		name       string
		cursor     int
		itemCount  int
		offset     int
		height     int
		wantCursor int
		wantOffset int
	}{
		{
			name:       "normal move down",
			cursor:     0,
			itemCount:  5,
			offset:     0,
			height:     10,
			wantCursor: 1,
			wantOffset: 0,
		},
		{
			name:       "at bottom of list",
			cursor:     4,
			itemCount:  5,
			offset:     0,
			height:     10,
			wantCursor: 4,
			wantOffset: 0,
		},
		{
			name:       "scroll needed",
			cursor:     9,
			itemCount:  20,
			offset:     0,
			height:     10,
			wantCursor: 10,
			wantOffset: 1,
		},
		{
			name:       "empty list",
			cursor:     0,
			itemCount:  0,
			offset:     0,
			height:     10,
			wantCursor: 0,
			wantOffset: 0,
		},
		{
			name:       "single item",
			cursor:     0,
			itemCount:  1,
			offset:     0,
			height:     10,
			wantCursor: 0,
			wantOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCursor, gotOffset := calcCursorDown(tt.cursor, tt.itemCount, tt.offset, tt.height)
			if gotCursor != tt.wantCursor || gotOffset != tt.wantOffset {
				t.Errorf("calcCursorDown() = (%d, %d), want (%d, %d)",
					gotCursor, gotOffset, tt.wantCursor, tt.wantOffset)
			}
		})
	}
}

func TestCalcCursorUp(t *testing.T) {
	tests := []struct {
		name       string
		cursor     int
		offset     int
		wantCursor int
		wantOffset int
	}{
		{
			name:       "normal move up",
			cursor:     3,
			offset:     0,
			wantCursor: 2,
			wantOffset: 0,
		},
		{
			name:       "at top of list",
			cursor:     0,
			offset:     0,
			wantCursor: 0,
			wantOffset: 0,
		},
		{
			name:       "scroll needed",
			cursor:     5,
			offset:     5,
			wantCursor: 4,
			wantOffset: 4,
		},
		{
			name:       "no scroll when cursor above offset",
			cursor:     3,
			offset:     2,
			wantCursor: 2,
			wantOffset: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCursor, gotOffset := calcCursorUp(tt.cursor, tt.offset)
			if gotCursor != tt.wantCursor || gotOffset != tt.wantOffset {
				t.Errorf("calcCursorUp() = (%d, %d), want (%d, %d)",
					gotCursor, gotOffset, tt.wantCursor, tt.wantOffset)
			}
		})
	}
}

func TestCalcRemoveItem(t *testing.T) {
	tests := []struct {
		name       string
		items      []note
		cursor     int
		wantLen    int
		wantCursor int
	}{
		{
			name:       "remove middle item",
			items:      []note{{title: "a"}, {title: "b"}, {title: "c"}},
			cursor:     1,
			wantLen:    2,
			wantCursor: 1,
		},
		{
			name:       "remove last item",
			items:      []note{{title: "a"}, {title: "b"}, {title: "c"}},
			cursor:     2,
			wantLen:    2,
			wantCursor: 1,
		},
		{
			name:       "remove first item",
			items:      []note{{title: "a"}, {title: "b"}, {title: "c"}},
			cursor:     0,
			wantLen:    2,
			wantCursor: 0,
		},
		{
			name:       "remove only item",
			items:      []note{{title: "a"}},
			cursor:     0,
			wantLen:    0,
			wantCursor: 0,
		},
		{
			name:       "empty list",
			items:      []note{},
			cursor:     0,
			wantLen:    0,
			wantCursor: 0,
		},
		{
			name:       "invalid cursor negative",
			items:      []note{{title: "a"}, {title: "b"}},
			cursor:     -1,
			wantLen:    2,
			wantCursor: -1,
		},
		{
			name:       "invalid cursor out of bounds",
			items:      []note{{title: "a"}, {title: "b"}},
			cursor:     5,
			wantLen:    2,
			wantCursor: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotItems, gotCursor := calcRemoveItem(tt.items, tt.cursor)
			if len(gotItems) != tt.wantLen || gotCursor != tt.wantCursor {
				t.Errorf("calcRemoveItem() len=%d, cursor=%d, want len=%d, cursor=%d",
					len(gotItems), gotCursor, tt.wantLen, tt.wantCursor)
			}
		})
	}
}

func TestCalcAddItem(t *testing.T) {
	tests := []struct {
		name       string
		itemCount  int
		offset     int
		height     int
		wantCursor int
		wantOffset int
	}{
		{
			name:       "fits within panel",
			itemCount:  3,
			offset:     0,
			height:     10,
			wantCursor: 2,
			wantOffset: 0,
		},
		{
			name:       "just overflows",
			itemCount:  11,
			offset:     0,
			height:     10,
			wantCursor: 10,
			wantOffset: 1,
		},
		{
			name:       "already scrolled and overflows",
			itemCount:  15,
			offset:     4,
			height:     10,
			wantCursor: 14,
			wantOffset: 5,
		},
		{
			name:       "single item from empty",
			itemCount:  1,
			offset:     0,
			height:     10,
			wantCursor: 0,
			wantOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCursor, gotOffset := calcAddItem(tt.itemCount, tt.offset, tt.height)
			if gotCursor != tt.wantCursor || gotOffset != tt.wantOffset {
				t.Errorf("calcAddItem() = (%d, %d), want (%d, %d)",
					gotCursor, gotOffset, tt.wantCursor, tt.wantOffset)
			}
		})
	}
}

func TestCalcRemoveItemImmutability(t *testing.T) {
	original := []note{{title: "a"}, {title: "b"}, {title: "c"}}
	originalLen := len(original)

	_, _ = calcRemoveItem(original, 1)

	if len(original) != originalLen {
		t.Errorf("original slice was mutated: len=%d, want=%d", len(original), originalLen)
	}
}
