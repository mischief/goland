package game

const (
	PropPos PropType = iota
	PropStaticSprite
	PropCamera
)

// A unique ID for a particular property.
type PropType uint

// All properties must be able to return a type.
type Property interface {
	Type() PropType
}
