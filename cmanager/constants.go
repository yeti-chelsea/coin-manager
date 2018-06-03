package cmanager

const (
	HELLO_MSG     string = "Hello"
	ACK_HELLO_MSG string = "Ack-Hello"
	KEEP_ALIVE    string = "Keep-Alive"
	SEND_BASIC    string = "Send-Basic"
	MINER_DAEMONS string = "Miner-Daemons"
	MINER_COINS   string = "Miner-Coins"
	HOST_NAME     string = "Host-Name"
)

const (
	CONSEQUTIVE_KEEP_ALIVE_TIMEOUT      int = 3
	CONSEQUTIVE_SEND_FAILURES           int = 3
	SLEEP_TIME_BEFORE_INTERACTING_INSEC int = 2
	TIMEOUT_BETWEEN_KEEP_ALIVE_INSEC    int = 4
	SEND_TIMEOUT_INSEC                  int = 1
	SLEEP_TIME_AFTER_KEEP_ALIVE_TIMEOUT int = 2
)

const (
	REGISTERED_MINER_IP		string = "miner-ip"
	REGISTERED_MINER_DAEMONS	string = "miner-daemons"
	REGISTERED_MINER_COINS		string = "miner-coins"
	REGISTERED_MINER_HOST		string = "miner-host"
)
