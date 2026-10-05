package models

// AccessScope describes the rows a logged-in user may access. Route
// permissions still decide which actions are available; this scope prevents
// horizontal access across requesters and stores.
type AccessScope struct {
	UserID     int
	StoreIDs   []int
	CanViewAll bool
}

func (s AccessScope) AllowsStore(storeID int) bool {
	if s.CanViewAll {
		return true
	}
	if storeID <= 0 {
		return false
	}
	for _, allowedID := range s.StoreIDs {
		if allowedID == storeID {
			return true
		}
	}
	return false
}

func (s AccessScope) AllowsOwnerOrStore(ownerUserID, storeID int) bool {
	return s.CanViewAll || (s.UserID > 0 && s.UserID == ownerUserID) || s.AllowsStore(storeID)
}
