package source

type AppID int32
type TheShip struct {
	Mode      int   `json:"Mode"`
	Witnesses uint8 `json:"Witnesses"`
	Duration  uint8 `json:"Duration"`
}

type ServerExtend struct {
	Port     uint16 `json:"Port"`
	SteamID  uint64 `json:"SteamID"`
	Keywords string `json:"Keywords"`
	GameID   uint64 `json:"GameID"`
}

type TV struct {
	Port uint16 `json:"Port"`
	Name string `json:"Name"`
}

type Mod struct {
	Link         string `json:"Link"`
	DownloadLink string `json:"DownloadLink"`
	Version      int32  `json:"Version"`
	Size         int32  `json:"Size"`
	Type         uint8  `json:"Type"`
	DLL          uint8  `json:"DLL"`
}

type Server struct {
	Proto        uint8         `json:"Protocol"`
	Name         string        `json:"Name"`
	Map          string        `json:"Map"`
	Folder       string        `json:"Folder"`
	Game         string        `json:"Game"`
	ID           int32         `json:"ID"`
	MaxPlayers   uint8         `json:"MaxPlayers"`
	PlayersCount uint8         `json:"PlayersCount"`
	Bots         uint8         `json:"Bots"`
	Type         string        `json:"Type"`
	System       string        `json:"System"`
	Visibility   bool          `json:"Visibility"`
	VAC          bool          `json:"VAC"`
	Ship         *TheShip      `json:"Ship,omitempty"`
	Version      string        `json:"Version"`
	Extend       *ServerExtend `json:"Extend,omitempty"`
	SourceTV     *TV           `json:"SourceTV,omitempty"`
	EDF          uint8         `json:"EDF,omitempty"`
	Mod          *Mod          `json:"Mod,omitempty"`
	Address      string        `json:"Address"`
}

func GetServerSystemString(os uint8) string {
	switch os {
	case uint8('L'), uint8('l'):
		return "Linux"
	case uint8('W'), uint8('w'):
		return "Windows"
	case uint8('m'):
		return "Mac"
	}

	return "Unknown"
}

func GetServerTypeString(st uint8) string {
	switch st {
	case uint8('d'), uint8('D'):
		return "Dedicated"
	case uint8('l'), uint8('L'):
		return "NonDedicated"
	case uint8('p'), uint8('P'):
		return "SourceTV"
	}

	return "Unknown"
}

type PingResponse struct {
	Address string `json:"address"`
	Status  bool   `json:"status"`
}
