package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/if41/steam"
)

func main() {
	steamLogin := flag.String("username", "", "Steam login")
	steamPassword := flag.String("password", "", "Steam password")
	steamSharedSecret := flag.String("shared-secret", "", "Steam shared secret")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	timeTip, err := steam.GetTimeTip()
	handlePanic(err)

	timeDiff := time.Duration(timeTip.Time - time.Now().Unix())
	session := steam.NewSession(&http.Client{}, "")
	handlePanic(session.Login(*steamLogin, *steamPassword, *steamSharedSecret, timeDiff))
	log.Print("Login successful")

	key, err := session.GetWebAPIKey()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Key: ", key)

	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			offers, err := session.GetTradeOffers(
				steam.TradeFilterActiveOnly|steam.TradeFilterUseTimeCutoff|steam.TradeFilterRecvOffers|steam.TradeFilterItemDescriptions,
				time.Now().Add(-time.Minute * 5),
			)

			if err != nil {
				log.Println(err)
				continue
			}

			for _, offer := range offers.ReceivedOffers {
				processOffer(session, offer)
			}
		}
	}
}

func processOffer(session *steam.Session, offer *steam.TradeOffer) {
	log.Printf("Offer id: %d, Offer state: %s", offer.ID, tradeStateDescriptions[offer.State])

	if offer.State == steam.TradeStateActive {
		for _, item := range offer.RecvItems {
			if item.Description != nil {
				log.Println(item.Description.Name)
			}
		}

		if err := session.AcceptTradeOffer(offer.ID); err != nil {
			log.Printf("error accepting offer: %v", err)
		} else {
			log.Println("accept request sent")
		}
	}
}

func handlePanic(err error) {
	if err != nil {
		panic(err)
	}
}

var tradeStateDescriptions = map[uint8]string{
	steam.TradeStateNone:                     "None",
	steam.TradeStateInvalid:                  "Invalid",
	steam.TradeStateActive:                   "Active",
	steam.TradeStateAccepted:                 "Accepted",
	steam.TradeStateCountered:                "Countered",
	steam.TradeStateExpired:                  "Expired",
	steam.TradeStateCanceled:                 "Canceled",
	steam.TradeStateDeclined:                 "Declined",
	steam.TradeStateInvalidItems:             "InvalidItems",
	steam.TradeStateCreatedNeedsConfirmation: "CreatedNeedsConfirmation",
	steam.TradeStateCanceledByTwoFactor:      "CanceledByTwoFactor",
	steam.TradeStateInEscrow:                 "InEscrow",
}
