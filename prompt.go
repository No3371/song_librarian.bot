package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"No3371.github.com/song_librarian.bot/binding"
	"No3371.github.com/song_librarian.bot/logger"
	"No3371.github.com/song_librarian.bot/redirect"
	"github.com/c-bata/go-prompt"
	"github.com/diamondburned/arikawa/v3/state"
)

const (
	commandBind = "bind"
	commandBindQuery = "bind_query"
	commandBindRemove = "bind_remove"
)

func promptLoop (s *state.State, ctx context.Context) {
	for (true) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		input := prompt.Input("", func (d prompt.Document) []prompt.Suggest {
			if strings.HasPrefix(d.Text, "bind") {
				return []prompt.Suggest {
					{
						Text:        commandBind,
						Description: "",
					},
					{
						Text:        commandBindQuery,
						Description: "",
					},
					{
						Text:        commandBindRemove,
						Description: "",
					},
				}
			}

			return []prompt.Suggest {
				{
					Text:        "add_channel",
					Description: "",
				},
				{
					Text:        "add_redirection",
					Description: "",
				},
				{
					Text:        "map",
					Description: "",
				},
				{
					Text:        "resetcommands",
					Description: "",
				},
			}
		})
		// if err != nil {
		// 	fmt.Printf("Prompt failed %v\n", err)
		// 	return
		// }

		err := handle(input, s)
		if err != nil {
			logger.Logger.Errorf("%v", err)
		}
	}
}

func handle (input string, s *state.State) (err error) {
	switch input {
	case commandBind:
		_, err = bind()
	case commandBindRemove:
		err = unbind()
	case commandBindQuery:
		break
	case "add_channel":
	case "add_redirection":
	case "map":
	case "resetcommands":
		resetAllCommands(s)
		break
	default:
		return errors.New("Unexpected command")
	}
	return nil
}

// bind:
func bind () (bId int, err error) {
	var cId uint64
	cId, err = enterUInt64("channel ID: ")
	if err != nil {
		return 0, err
	}
	
	enteringBId:
	input := prompt.Input("binding ID, or 'new': ", noopCompleter)
	
	if input == "cancel" {
		return 0, errors.New("cancelled")
	}

	if input == "new" {
		bId, err = newBinding()
		fmt.Printf("Created binding#%d\n", bId)
		b := binding.GetModifiableBinding(bId)
		fmt.Println("Creating fist redirection...")
		
		r, rcId, err := newRedirection()
		if err != nil {
			return 0, err
		}
		
		b.SetRedirection(r, rcId)
		b.EnableUrlRegexes(0)
	
		binding.Bind(cId, bId)
	} else {
		bId, err = strconv.Atoi(input)
		fmt.Printf("Getting binding#%d\n", bId)
		b := binding.GetModifiableBinding(bId)	
		if b == nil {
			fmt.Printf("Binding#%d does not exist.\n", bId)
			goto enteringBId
		}
		logger.Logger.Infof("%+v", b.ChannelBinding)	
		r, rcId, err := newRedirection()
		if err != nil {
			return 0, err
		}		
		b.SetRedirection(r, rcId)	
		binding.Bind(cId, bId)
		logger.Logger.Infof("%+v", b.ChannelBinding)	
	}
	
	binding.SaveAll()
	if err != nil {
		return 0, err
	}

	return bId, err
	
}

const (
	selectionOriginalSong = "original_songs"
	selectionCoverSong = "cover_songs"
	selectionSingingStream = "singing_stream"
)

func chooseRedirect () redirect.RedirectType {
	choose: switch prompt.Choose("Redirection type: ", []string { selectionOriginalSong, selectionCoverSong, selectionSingingStream, "cancel" }) {
	case selectionOriginalSong:
		return redirect.Original
	case selectionCoverSong:
		return redirect.Cover
	case selectionSingingStream:
		return redirect.Stream
	case "cancel":
		return redirect.None
	default:
		fmt.Println("Unexpected input, choose from 3 valid options or 'cancel'")
		goto choose
	}
}

func enterUInt64 (prefix string) (uint64, error) {
	input: _num := prompt.Input(prefix, noopCompleter)
	if _num == "cancel" {
		return 0, errors.New("cancelled")
	}
	if len(_num) == 0 {
		fmt.Printf("Empty input, re-enter valid value or 'cancel'")
		goto input
	}
	num, err := strconv.ParseUint(_num, 10, 64)
	if err != nil {
		return 0, err
	}

	return num, nil
}

func enterInt (prefix string) (int, error) {
	input: _num := prompt.Input(prefix, noopCompleter)
	if _num == "cancel" {
		return 0, errors.New("cancelled")
	}
	if len(_num) == 0 {
		fmt.Printf("Empty input, re-enter valid value or 'cancel'")
		goto input
	}
	num, err := strconv.Atoi(_num)
	if err != nil {
		return 0, err
	}

	return num, nil
}

func unbind () (err error) {
	var cId uint64
	var bId int
	cId, err = enterUInt64("Channel ID: ")
	if err != nil {
		return err
	}
	bId, err = enterInt("Binding ID: ")
	if err != nil {
		return err
	}

	binding.Unbind(cId, bId)

	binding.SaveAll()

	return nil
}

func queryBind () (err error) {
	var bId int
	bId, err = enterInt("Binding ID: ")
	if err != nil {
		return err
	}

	b := binding.QueryBinding(bId)
	if b == nil {
		fmt.Printf("Binding#%d not exist\n", bId)
		return nil
	}

	fmt.Printf("%+v", b)
	return nil
}

func newBinding () (int, error) {
	bId := binding.NewBinding()
	return bId, nil
}

func newRedirection () (redirect.RedirectType, uint64, error) {
	r := chooseRedirect()
	
	cId, err := enterUInt64("Dest channel ID: ")
	if err != nil {
		return 0, 0, err
	}

	return r, cId, nil

}

var emptySuggestList []prompt.Suggest = make([]prompt.Suggest, 0)
func noopCompleter (prompt.Document) []prompt.Suggest {
	return emptySuggestList
}