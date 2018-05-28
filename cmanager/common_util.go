package cmanager

type MIps struct {
	Ips []string `json:REGISTERED_MINER_IP`
}

type MDaemons struct {
	Daemons []string `json:REGISTERED_MINER_DAEMONS`
}

type MCoins struct {
	Coins []string `json:REGISTERED_MINER_COINS`
}
