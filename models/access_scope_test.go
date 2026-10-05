package models

import "testing"

func TestAccessScopeFailsClosedWithoutStoreMapping(t *testing.T) {
	scope := AccessScope{UserID: 10}
	if scope.AllowsStore(5) {
		t.Fatal("unmapped store must not be accessible")
	}
	if !scope.AllowsOwnerOrStore(10, 5) {
		t.Fatal("request owner should retain access to their own request")
	}
}

func TestAccessScopeAllowsMappedAndGlobalStores(t *testing.T) {
	mapped := AccessScope{UserID: 10, StoreIDs: []int{2, 5}}
	if !mapped.AllowsStore(5) || mapped.AllowsStore(6) {
		t.Fatal("mapped scope returned an unexpected store decision")
	}

	global := AccessScope{UserID: 1, CanViewAll: true}
	if !global.AllowsStore(999) {
		t.Fatal("global scope should allow every valid store")
	}
}
