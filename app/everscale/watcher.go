package everscale

import (
	"encoding/json"
	"fmt"
	goton "github.com/move-ton/ton-client-go"
	"github.com/move-ton/ton-client-go/domain"
	"log"
	"time"
)

type EventContractWatcher struct {
	Ton     *goton.Ton
	ABI     *domain.Abi
	Address string
	Cache   map[string]struct{}
}

func New(address string, ton *goton.Ton, abi *domain.Abi) *EventContractWatcher {
	return &EventContractWatcher{ton, abi, address, make(map[string]struct{})}
}

func (cw *EventContractWatcher) RunWatcher(channel chan *domain.DecodedMessageBody) {
	// It's place for subscription, but it doesn't work
	count := 0
	for {
		collection, err := cw.Ton.Net.QueryCollection(&domain.ParamsOfQueryCollection{
			Collection: "messages",
			Filter:     json.RawMessage(fmt.Sprintf(`{"src":{"eq":"%s"}, "msg_type":{"eq":2}}`, cw.Address)),
			Result:     "id, created_lt, body",
			Order:      []*domain.OrderBy{{"created_lt", "ASC"}},
		})
		if err != nil {
			log.Println(err)
			return
		}

		for _, elem := range collection.Result {
			var msg struct {
				ID   string `json:"id"`
				Body string `json:"body"`
			}
			err := json.Unmarshal(elem, &msg)
			if err != nil {
				log.Println(err)
			}
			if count == 0 {
				cw.Cache[msg.ID] = struct{}{}
			}
			if _, ok := cw.Cache[msg.ID]; !ok {
				cw.Cache[msg.ID] = struct{}{}

				res, err := cw.Ton.Abi.DecodeMessageBody(&domain.ParamsOfDecodeMessageBody{
					Body: msg.Body,
					Abi:  cw.ABI,
				})

				if err != nil {
					log.Println(err)
				}

				channel <- res
			}
		}
		if count == 0 {
			count += 1
		}

		time.Sleep(5 * time.Second)
	}
}
