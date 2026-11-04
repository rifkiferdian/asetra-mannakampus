package models

// Asset lifecycle values live in one place so controllers, services, and
// repositories do not depend on duplicated string literals.
const (
	AssetStatusAvailable   = "AVAILABLE"
	AssetStatusInUse       = "IN_USE"
	AssetStatusMaintenance = "MAINTENANCE"
	AssetStatusBroken      = "BROKEN"
	AssetStatusDisposed    = "DISPOSED"

	ComponentStatusInStorage   = "IN_STORAGE"
	ComponentStatusInstalled   = "INSTALLED"
	ComponentStatusMaintenance = "MAINTENANCE"
	ComponentStatusBroken      = "BROKEN"
	ComponentStatusDisposed    = "DISPOSED"

	AssetMovementReceive     = "RECEIVE"
	AssetMovementTransfer    = "TRANSFER"
	AssetMovementAssign      = "ASSIGN"
	AssetMovementReturn      = "RETURN"
	AssetMovementMaintenance = "MAINTENANCE"
	AssetMovementBroken      = "BROKEN"

	ComponentMovementReceive         = "RECEIVE"
	ComponentMovementInstall         = "INSTALL"
	ComponentMovementUninstall       = "UNINSTALL"
	ComponentMovementTransfer        = "TRANSFER"
	ComponentMovementMaintenance     = "MAINTENANCE"
	ComponentMovementReturnToStorage = "RETURN_TO_STORAGE"
	ComponentMovementBroken          = "BROKEN"
	ComponentMovementDispose         = "DISPOSE"
)
