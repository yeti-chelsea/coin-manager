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

type AllMinerInfo struct {
	Ip string `json:REGISTERED_MINER_IP`
	Daemons []string `json:REGISTERED_MINER_DAEMONS`
	Coins []string `json:REGISTERED_MINER_COINS`
}
