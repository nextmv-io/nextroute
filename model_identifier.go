// © 2019-present nextmv.io inc

package nextroute

// Identifier is an interface that can be used for identifying objects.
type Identifier interface {
	// ID returns the identifier of the object.
	ID() string
	// SetID sets the identifier of the object.
	SetID(string)
}
