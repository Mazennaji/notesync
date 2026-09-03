package sync

type Action string

const (
	Skip     Action = "skip"
	Push     Action = "push"
	Pull     Action = "pull"
	Conflict Action = "conflict"
	NewLocal Action = "new_local"
)

type Decision struct {
	NotePath   string
	Action     Action
	LocalHash  string
	RemoteHash string
}

func Classify(notePath, localHash, remoteHash, lastSyncedHash string) Decision {
	d := Decision{NotePath: notePath, LocalHash: localHash, RemoteHash: remoteHash}

	if lastSyncedHash == "" {
		d.Action = NewLocal
		return d
	}

	localChanged := localHash != lastSyncedHash
	remoteChanged := remoteHash != lastSyncedHash

	switch {
	case !localChanged && !remoteChanged:
		d.Action = Skip
	case localChanged && !remoteChanged:
		d.Action = Push
	case !localChanged && remoteChanged:
		d.Action = Pull
	default:
		d.Action = Conflict
	}
	return d
}
