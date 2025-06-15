package p2p

type HandShakeFunc func(Peer) error

func DefaultHandShake(peer Peer) error {
	peer.SetID(peer.RemoteAddr().String())
	return nil
}

func PSKAndECDHEHandshake(peer Peer) error {
	/**
	  * will recieve the first packet from peer and will get the PSK and public key,
	* then it will verify the psk from env variable/config, then will create
	* ecdhe asymetric keys and will share the number of peers connected along with
	* the public key, will store the peers public key and our private key.
	  * **/

	return nil
}
