package store

type ResyncOperation uint

const (
	RESYNCFILE ResyncOperation = 0
	RESYNCMEM  ResyncOperation = 1
)

type OPERATION uint

const (
	NOK OPERATION = 0 // not ok

	// Need to send chunk
	NTS    OPERATION = 1
	NTS_OK OPERATION = 2

	// WRITE
	WRITE  = 3 // single Peer
	RESYNC = 4 // Broadcast

)
