package common

type Direction int

const (
	IconToFrame Direction = iota
	FrameToIcon Direction = iota
)

func (d Direction) String() string {
	if d == IconToFrame {
		return "Icon->Frame"
	} else {
		return "Frame->Icon"
	}
}
