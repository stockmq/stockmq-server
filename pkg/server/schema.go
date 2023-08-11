package server

// Header represents HTTP header
type Header struct {
	Name string `xml:"Name,attr"`
	Text string `xml:",chardata"`
}

// MessageHeader represents common fields for each message.
type MessageHeader struct {
	Symbol  string `json:"symbol"`
	Source  string `json:"source"`
	Time    int64  `json:"time"`
	TimeSrv int64  `json:"time_srv"`
	TimeRcv int64  `json:"time_rcv"`
}

// Candle represents OLHCV bar.
type Candle struct {
	MessageHeader

	Interval string `json:"interval"`
	Open     string `json:"open"`
	High     string `json:"high"`
	Low      string `json:"low"`
	Close    string `json:"close"`
	Volume   string `json:"volume"`
}

// Quote represents bid and ask
type Quote struct {
	MessageHeader

	BidsDepth int        `json:"bids_depth"`
	Bids      [][]string `json:"bids"`
	AsksDepth int        `json:"asks_depth"`
	Asks      [][]string `json:"asks"`
}
