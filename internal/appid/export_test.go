package appid

// The a11y registry-interest corpus rides the testpackage rule (external
// test package): the unexported seam and its wire constants surface here —
// the aliases never leak into the production API.

//nolint:gochecknoglobals // the test-only seam aliases (the wire-form precedent)
var (
	RegisterEventInterest = registerEventInterest
	EventStateInterest    = eventStateInterest
	A11yRegistryName      = a11yRegistryName
	BridgeApp             = bridgeApp
)
