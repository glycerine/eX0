#pragma once
#ifndef __NetworkStateAuther_H__
#define __NetworkStateAuther_H__

class NetworkStateAuther
	: public PlayerStateAuther
{
public:
	NetworkStateAuther(CPlayer & oPlayer);
	~NetworkStateAuther();

	void ProcessCommands();
	void ProcessWpnCommands();
	void ProcessUpdates();
	void SendUpdate();

	bool IsLocal(void) { return false; }

	static void ProcessUpdate(CPacket & oPacket);

	//u_char		cLastAckedCommandSequenceNumber;
	u_char		cCurrentCommandSeriesNumber;
};

#endif // __NetworkStateAuther_H__
