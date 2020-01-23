package ccache

type eventKind int32

const (
	eventKindDelete eventKind = iota
	eventKindUpdate
)

type event struct {
	item *Item
	kind eventKind
}
