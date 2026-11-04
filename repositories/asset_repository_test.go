package repositories

import (
	"gobase-app/models"
	"reflect"
	"testing"
)

func TestAssetStoreScopeSQL(t *testing.T) {
	tests := []struct {
		name      string
		scope     models.AccessScope
		wantWhere string
		wantArgs  []any
	}{
		{name: "global", scope: models.AccessScope{CanViewAll: true}, wantWhere: "1=1"},
		{name: "no stores", scope: models.AccessScope{}, wantWhere: "1=0"},
		{name: "assigned stores", scope: models.AccessScope{StoreIDs: []int{2, 7}}, wantWhere: "a.store_id IN (?,?)", wantArgs: []any{2, 7}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			where, args := assetStoreScopeSQL("a.store_id", test.scope)
			if where != test.wantWhere {
				t.Fatalf("where = %q, want %q", where, test.wantWhere)
			}
			if !reflect.DeepEqual(args, test.wantArgs) {
				t.Fatalf("args = %#v, want %#v", args, test.wantArgs)
			}
		})
	}
}
