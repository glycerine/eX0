package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/shurcooL/eX0/eX0-go/packet"
	"github.com/shurcooL/go-goon"
)

var tcp, udp net.Conn
var lastAckedCmdSequenceNumber uint8

func connectToServer(addr string, c *character) {
	var err error
	tcp, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	//defer tcp.Close()

	var signature = uint64(time.Now().UnixNano())

	{
		var p = packet.JoinServerRequest{
			TcpHeader: packet.TcpHeader{
				Type: packet.JoinServerRequestType,
			},
			Version:    1,
			Passphrase: [16]byte{'s', 'o', 'm', 'e', 'r', 'a', 'n', 'd', 'o', 'm', 'p', 'a', 's', 's', '0', '1'},
			Signature:  signature,
		}

		p.Length = 26

		err := binary.Write(tcp, binary.BigEndian, &p)
		if err != nil {
			panic(err)
		}
	}

	{
		var r packet.JoinServerAccept
		err := binary.Read(tcp, binary.BigEndian, &r)
		if err != nil {
			panic(err)
		}
		goon.Dump(r)

		state.TotalPlayerCount = r.TotalPlayerCount + 1
	}

	udp, err = net.Dial("udp", addr)
	if err != nil {
		panic(err)
	}

	{
		var p packet.Handshake
		p.Type = packet.HandshakeType
		p.Signature = signature

		err := binary.Write(udp, binary.BigEndian, &p)
		if err != nil {
			panic(err)
		}
	}

	{
		var r packet.UdpConnectionEstablished
		err := binary.Read(tcp, binary.BigEndian, &r)
		if err != nil {
			panic(err)
		}
		goon.Dump(r)
	}

	{
		const name = "shurcooL"

		var p packet.LocalPlayerInfo
		p.Type = packet.LocalPlayerInfoType
		p.NameLength = uint8(len(name))
		p.Name = []byte(name)
		p.CommandRate = 20
		p.UpdateRate = 20

		p.Length = 3 + uint16(len(name))

		err := binary.Write(tcp, binary.BigEndian, &p.TcpHeader)
		if err != nil {
			panic(err)
		}
		err = binary.Write(tcp, binary.BigEndian, &p.NameLength)
		if err != nil {
			panic(err)
		}
		err = binary.Write(tcp, binary.BigEndian, &p.Name)
		if err != nil {
			panic(err)
		}
		err = binary.Write(tcp, binary.BigEndian, &p.CommandRate)
		if err != nil {
			panic(err)
		}
		err = binary.Write(tcp, binary.BigEndian, &p.UpdateRate)
		if err != nil {
			panic(err)
		}
	}

	{
		var r packet.LoadLevel
		err := binary.Read(tcp, binary.BigEndian, &r.TcpHeader)
		if err != nil {
			panic(err)
		}
		r.LevelFilename = make([]byte, r.Length)
		err = binary.Read(tcp, binary.BigEndian, &r.LevelFilename)
		if err != nil {
			panic(err)
		}
		goon.Dump(r)
		goon.Dump(string(r.LevelFilename))
	}

	{
		var r packet.CurrentPlayersInfo
		err := binary.Read(tcp, binary.BigEndian, &r.TcpHeader)
		if err != nil {
			panic(err)
		}
		r.Players = make([]packet.PlayerInfo, state.TotalPlayerCount)
		for i := range r.Players {
			var playerInfo packet.PlayerInfo
			err = binary.Read(tcp, binary.BigEndian, &playerInfo.NameLength)
			if err != nil {
				panic(err)
			}

			if playerInfo.NameLength != 0 {
				playerInfo.Name = make([]byte, playerInfo.NameLength)
				err = binary.Read(tcp, binary.BigEndian, &playerInfo.Name)
				if err != nil {
					panic(err)
				}

				err = binary.Read(tcp, binary.BigEndian, &playerInfo.Team)
				if err != nil {
					panic(err)
				}

				if playerInfo.Team != 2 {
					playerInfo.State = new(packet.State)
					err = binary.Read(tcp, binary.BigEndian, playerInfo.State)
					if err != nil {
						panic(err)
					}
				}
			}

			r.Players[i] = playerInfo
		}
		goon.Dump(r)
	}

	{
		var r packet.EnterGamePermission
		err := binary.Read(tcp, binary.BigEndian, &r)
		if err != nil {
			panic(err)
		}
		goon.Dump(r)
	}

	{
		var p packet.EnteredGameNotification
		p.Type = packet.EnteredGameNotificationType

		p.Length = 0

		err := binary.Write(tcp, binary.BigEndian, &p)
		if err != nil {
			panic(err)
		}
	}

	time.Sleep(3 * time.Second)

	{
		var p packet.JoinTeamRequest
		p.Type = packet.JoinTeamRequestType
		p.Team = 0

		p.Length = 1

		err := binary.Write(tcp, binary.BigEndian, &p.TcpHeader)
		if err != nil {
			panic(err)
		}
		err = binary.Write(tcp, binary.BigEndian, &p.Team)
		if err != nil {
			panic(err)
		}
	}

	{
		var r packet.PlayerJoinedTeam
		err := binary.Read(tcp, binary.BigEndian, &r.TcpHeader)
		if err != nil {
			panic(err)
		}
		err = binary.Read(tcp, binary.BigEndian, &r.PlayerId)
		if err != nil {
			panic(err)
		}
		err = binary.Read(tcp, binary.BigEndian, &r.Team)
		if err != nil {
			panic(err)
		}
		if r.Team != 2 {
			r.State = new(packet.State)
			err = binary.Read(tcp, binary.BigEndian, r.State)
			if err != nil {
				panic(err)
			}
		}
		goon.Dump(r)

		lastAckedCmdSequenceNumber = r.State.CommandSequenceNumber
	}

	fmt.Println("done")

	go func() {
		defer udp.Close()

		for {
			var b [packet.MAX_UDP_SIZE]byte
			n, err := udp.Read(b[:])
			if err != nil {
				panic(err)
			}
			var buf = bytes.NewReader(b[:n])

			var udpHeader packet.UdpHeader
			err = binary.Read(buf, binary.BigEndian, &udpHeader)
			if err != nil {
				panic(err)
			}

			switch udpHeader.Type {
			case packet.PingType:
				var r packet.Ping
				err = binary.Read(buf, binary.BigEndian, &r.PingData)
				if err != nil {
					panic(err)
				}
				r.LastLatencies = make([]uint16, state.TotalPlayerCount)
				err = binary.Read(buf, binary.BigEndian, &r.LastLatencies)
				if err != nil {
					panic(err)
				}

				{
					var p packet.Pong
					p.Type = packet.PongType
					p.PingData = r.PingData

					err := binary.Write(udp, binary.BigEndian, &p)
					if err != nil {
						panic(err)
					}
				}
			case packet.ServerUpdateType:
				var r packet.ServerUpdate
				err = binary.Read(buf, binary.BigEndian, &r.CurrentUpdateSequenceNumber)
				if err != nil {
					panic(err)
				}
				r.Players = make([]packet.PlayerUpdate, state.TotalPlayerCount)
				for i := range r.Players {
					var playerUpdate packet.PlayerUpdate
					err = binary.Read(buf, binary.BigEndian, &playerUpdate.ActivePlayer)
					if err != nil {
						panic(err)
					}

					if playerUpdate.ActivePlayer != 0 {
						playerUpdate.State = new(packet.State)
						err = binary.Read(buf, binary.BigEndian, playerUpdate.State)
						if err != nil {
							panic(err)
						}
					}

					r.Players[i] = playerUpdate
				}

				if playerUpdate := r.Players[0]; playerUpdate.ActivePlayer != 0 {
					c.pos[0] = playerUpdate.State.X
					c.pos[1] = playerUpdate.State.Y
					c.Z = playerUpdate.State.Z
				}
			}
		}
	}()

	//time.Sleep(10 * time.Second)
}
