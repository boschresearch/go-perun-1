// Copyright 2022 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protobuf

import (
	"time"

	"perun.network/go-perun/wire"
)

func fromPingMsg(msg *wire.PingMsg) (*Envelope_PingMsg, error) {
	return &Envelope_PingMsg{
		PingMsg: &PingMsg{
			Created: msg.Created.UnixNano(),
		},
	}, nil
}

func fromPongMsg(msg *wire.PongMsg) (*Envelope_PongMsg, error) {
	return &Envelope_PongMsg{
		PongMsg: &PongMsg{
			Created: msg.Created.UnixNano(),
		},
	}, nil
}

func fromShutdownMsg(msg *wire.ShutdownMsg) (*Envelope_ShutdownMsg, error) {
	return &Envelope_ShutdownMsg{
		ShutdownMsg: &ShutdownMsg{
			Reason: msg.Reason,
		},
	}, nil
}

func toPingMsg(protoMsg *Envelope_PingMsg) *wire.PingMsg {
	return &wire.PingMsg{
		PingPongMsg: wire.PingPongMsg{
			Created: time.Unix(0, protoMsg.PingMsg.Created),
		},
	}
}

func toPongMsg(protoMsg *Envelope_PongMsg) *wire.PongMsg {
	return &wire.PongMsg{
		PingPongMsg: wire.PingPongMsg{
			Created: time.Unix(0, protoMsg.PongMsg.Created),
		},
	}
}

func toShutdownMsg(protoMsg *Envelope_ShutdownMsg) *wire.ShutdownMsg {
	return &wire.ShutdownMsg{
		Reason: protoMsg.ShutdownMsg.Reason,
	}
}
