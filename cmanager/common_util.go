package cmanager

type MIps struct {
	Ips []string `json:"miner-ip"`
}

type MHosts struct {
	HostName []string `json:"miner-host"`
}

type MDaemons struct {
	Daemons []string `json:"miner-daemons"`
}

type MCoins struct {
	Coins []string `json:"miner-coins"`
}

type AllMinerInfo struct {
	Ip string `json:"miner-ip"`
	Daemons []string `json:"miner-daemons"`
	Coins []string `json:"miner-coins"`
}
