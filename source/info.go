package source

import (
	"errors"
	"fmt"
	"icraft/valve/packet"
)

// Info :get server info.
func (c Client) Info() (*Server, error) {
	var b packet.Builder
	b.WriteBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x54})
	b.WriteCString("Source Engine Query")

	if err := c.send(b.Bytes()); err != nil {
		return nil, err
	}

	bytes, err := c.receive()
	if err != nil {
		return nil, err
	}

	reader := packet.NewReader(bytes)
	header := reader.ReadInt32()
	if header != -1 {
		return nil, errors.New("source.Client.Info: packet header mismatch")
	}

	protocol := reader.ReadUint8()
	switch protocol {
	case uint8('m'): // Obsolete GoldSource Response
		srv := ResolveObsoleteGoldSourceResponse(reader)
		srv.Address = c.addr

		return srv, err
	case uint8('l'): // Latest Version Source Response
		server := ResolveSourceResponse(reader)
		server.Address = c.addr

		return server, nil
	}

	return nil, errors.New(fmt.Sprintf("source.Client.Info: unsupported protocol, header: %d, protocol: %x", header, protocol))
}

func ResolveSourceResponse(r *packet.Reader) *Server {
	srv := &Server{
		Name:         r.ReadString(),
		Map:          r.ReadString(),
		Folder:       r.ReadString(),
		Game:         r.ReadString(),
		ID:           r.ReadInt32(), // TODO: fix it.
		PlayersCount: r.ReadUint8(),
		MaxPlayers:   r.ReadUint8(),
		Bots:         r.ReadUint8(),
		Type:         GetServerTypeString(r.ReadUint8()),
		System:       GetServerSystemString(r.ReadUint8()),
		Visibility:   r.ReadUint8() == 1,
		VAC:          r.ReadUint8() == 1,
	}

	// TheShip:
	if srv.ID == 2400 {
		srv.Ship = &TheShip{
			Mode: int(r.ReadUint8()), // TODO: fix it.
		}
	} else {
		srv.Version = r.ReadString()
	}

	// Has More
	if !r.More() {
		return srv
	}

	srv.EDF = r.ReadUint8()
	if (srv.EDF & 0x80) != 0 {
		srv.Extend.Port = r.ReadUint16()
	}

	if (srv.EDF & 0x10) != 0 {
		srv.Extend.SteamID = r.ReadUint64()
	}

	if (srv.EDF & 0x40) != 0 {
		srv.SourceTV = &TV{}
		srv.SourceTV.Port = r.ReadUint16()
		srv.SourceTV.Name = r.ReadString()
	}

	if (srv.EDF & 0x20) != 0 {
		srv.Extend.Keywords = r.ReadString()
	}

	if (srv.EDF & 0x01) != 0 {
		srv.Extend.GameID = r.ReadUint64()
	}

	return srv
}

func ResolveObsoleteGoldSourceResponse(r *packet.Reader) *Server {
	r.ReadString()
	srv := &Server{
		Name:         r.ReadString(),
		Map:          r.ReadString(),
		Folder:       r.ReadString(),
		Game:         r.ReadString(),
		PlayersCount: r.ReadUint8(),
		MaxPlayers:   r.ReadUint8(),
		Proto:        r.ReadUint8(),
		Type:         GetServerTypeString(r.ReadUint8()),
		System:       GetServerSystemString(r.ReadUint8()),
		Visibility:   r.ReadUint8() == 1,
	}

	mod := r.ReadUint8()
	if mod == 1 {
		srv.Mod = &Mod{Link: r.ReadString(), DownloadLink: r.ReadString()}
		r.ReadUint8()
		srv.Mod.Version = r.ReadInt32()
		srv.Mod.Size = r.ReadInt32()
		srv.Mod.Type = r.ReadUint8()
		srv.Mod.DLL = r.ReadUint8()
	} else {
		srv.VAC = r.ReadUint8() == 1
		srv.Bots = r.ReadUint8()
	}

	return srv
}
