// Copyright 2020 Antrea Authors
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

package openflow

import (
	"time"

	"github.com/contiv/ofnet/ofctrl"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

type ofpPacketInReason uint

type PacketInHandler interface {
	HandlePacketIn(pktIn *ofctrl.PacketIn) error
}

const (
	// Action explicitly output to controller.
	ofprAction ofpPacketInReason = 1
)

func (c *client) RegisterPacketInHandler(packetHandlerName string, packetInHandler interface{}) {
	handler, ok := packetInHandler.(PacketInHandler)
	if !ok {
		klog.Errorf("Invalid Traceflow controller.")
		return
	}
	c.packetInHandlers[packetHandlerName] = handler
}

func (c *client) StartPacketInHandler(stopCh <-chan struct{}) {
	if len(c.packetInHandlers) == 0 {
		return
	}
	ch := make(chan *ofctrl.PacketIn)
	err := c.SubscribePacketIn(uint8(ofprAction), ch)
	if err != nil {
		klog.Errorf("Subscribe PacketIn failed %+v", err)
	}

	wait.PollUntil(time.Second, func() (done bool, err error) {
		pktIn := <-ch
		for name, handler := range c.packetInHandlers {
			err = handler.HandlePacketIn(pktIn)
			if err != nil {
				klog.Errorf("PacketIn handler %s failed to process packet: %+v", name, err)
			}
		}
		return false, err
	}, stopCh)
}
