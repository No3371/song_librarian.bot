package main

import (
	"github.com/diamondburned/arikawa/v3/discord"
)

var subscribingStates map[discord.UserID]bool

func init () {
	subscribingStates = make(map[discord.UserID]bool)
}

func unsub (u discord.UserID) (err error) {
	subscribingStates[u] = false
	return sv.SaveSubState(uint64(u), false)
}

func resub (u discord.UserID) (err error) {
	if _, loaded := subscribingStates[u]; !loaded {
		subscribingStates[u], err = sv.LoadSubState(uint64(u))
	}
	subscribingStates[u] = true
	return sv.SaveSubState(uint64(u), true)
}

func getSubState (u discord.UserID) (sub bool, err error) {
	if s, tracked := subscribingStates[u]; !tracked {
		s, err = sv.LoadSubState(uint64(u))
		if err != nil {
			return true, err
		}
		if s { // Not saved in database, default to true
			return true, nil
		} else {
			subscribingStates[u] = s
			return s, nil
		}
	} else {
		return s, nil
	}
}
