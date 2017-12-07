// Package order manages the bookkeeping and utilies required
// for users to create an 'order' meaning they have requested
// delegations for a certain resource.
//
// Copyright (c) 2016 CloudFlare, Inc.
package order

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/cloudflare/redoctober/notifier"
)

const (
	NewOrder       = "%s has created an order for the label %s. requesting %d delegations for %s"
	NewOrderLink   = "@%s - https://%s?%s"
	OrderFulfilled = "%s has had order %s fulfilled."
	NewDelegation  = "%s has delegated the label %s to %s (per order %s) for %s"
)

type Order struct {
	Creator string
	Users   []string
	Num     string

	TimeRequested     time.Time
	DurationRequested time.Duration
	Delegated         int
	OwnersDelegated   []string
	Owners            []string
	Labels            []string
}

type OrderIndex struct {
	OrderFor string

	OrderId     string
	OrderOwners []string
}

// Orders represents a mapping of Order IDs to Orders. This structure
// is useful for looking up information about individual Orders and
// whether or not an order has been fulfilled. Orders that have been
// fulfilled will be removed from the structure.
type Orderer struct {
	Orders        map[string]Order
	Notifier      notifier.Notifier
	RoHost        string
	AlternateName string
}

func CreateOrder(name, orderNum string, time time.Time, duration time.Duration, adminsDelegated, contacts, users, labels []string, numDelegated int) (ord Order) {
	ord.Creator = name
	ord.Num = orderNum
	ord.Labels = labels
	ord.TimeRequested = time
	ord.DurationRequested = duration
	ord.OwnersDelegated = adminsDelegated
	ord.Owners = contacts
	ord.Delegated = numDelegated
	ord.Users = users
	return
}

func GenerateNum() (num string) {
	b := make([]byte, 12)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// NewOrder will create a new map of Orders
func NewOrderer(roHost string, notifier notifier.Notifier) (o Orderer) {
	o.Orders = make(map[string]Order)
	o.Notifier = notifier
	o.RoHost = roHost
	o.AlternateName = "HipchatName"
	return
}

// notify is a generic function for using a notifier, but it checks to make
// sure that there is a notifier available, since there won't always be.
func notify(o *Orderer, msg, color string) (err error) {
	if o.Notifier != nil {
		err = o.Notifier.Notify(msg, color)
	} else {
		err = errors.New("no Notifier set to notify of order")
	}
	return
}

func (o *Orderer) NotifyNewOrder(duration, orderNum string, names, labels []string, uses int, owners map[string]string) error {
	labelList := ""
	for i, label := range labels {
		if i == 0 {
			labelList += label
		} else {
			// Never include spaces in something go URI encodes. Go will
			// add a + to the string, instead of a %20
			labelList += "," + label
		}
	}
	nameList := ""
	for i, name := range names {
		if i == 0 {
			nameList += name
		} else {
			// Never include spaces in something go URI encodes. Go will
			// add a + to the string, instead of a %20
			nameList += "," + name
		}
	}

	n := fmt.Sprintf(NewOrder, nameList, labelList, uses, duration)
	err := notify(o, n, notifier.RedBackground)
	if err != nil {
		return err
	}

	for owner, hipchatName := range owners {
		queryParams := url.Values{
			"delegator": {owner},
			"label":     {labelList},
			"duration":  {duration},
			"uses":      {strconv.Itoa(uses)},
			"ordernum":  {orderNum},
			"delegatee": {nameList},
		}.Encode()
		err = notify(o, fmt.Sprintf(NewOrderLink, hipchatName, o.RoHost, queryParams), notifier.GreenBackground)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Orderer) NotifyDelegation(delegator, delegatee, orderNum, duration string, labels []string) error {
	labelList := ""
	for i, label := range labels {
		if i == 0 {
			labelList += label
		} else {
			labelList += ", " + label
		}
	}
	n := fmt.Sprintf(NewDelegation, delegator, labelList, delegatee, orderNum, duration)
	return notify(o, n, notifier.YellowBackground)
}
func (o *Orderer) NotifyOrderFulfilled(name, orderNum string) error {
	n := fmt.Sprintf(OrderFulfilled, name, orderNum)
	return notify(o, n, notifier.PurpleBackground)
}

func (o *Orderer) FindOrder(user string, labels []string) (string, bool) {
	for key, order := range o.Orders {
		foundLabel := false
		foundUser := false
		for _, orderUser := range order.Users {
			if orderUser == user {
				foundUser = true
			}
		}
		if !foundUser {
			continue
		}
		for _, ol := range order.Labels {
			foundLabel = false
			for _, il := range labels {
				if il == ol {
					foundLabel = true
				}
			}
		}
		if !foundLabel {
			continue
		}
		return key, true
	}
	return "", false
}
