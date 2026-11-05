package game

import (
	"fmt"
	"strings"
	"mud/world"
)

type SubmenuCreateMode int
const (
	SUBMENU_CREATE_MODE_NAME SubmenuCreateMode = iota
	SUBMENU_CREATE_MODE_RACE
	SUBMENU_CREATE_MODE_CLASS
	SUBMENU_CREATE_MODE_SUMMARY
)

type SubmenuCreate struct {
	mode SubmenuCreateMode

	characterName string
	characterRace world.RaceId
	characterClass world.ClassId
}

func (submenu *SubmenuCreate) handleCommand(gamestate *GameState, player *Player, input string) bool {
	if strings.EqualFold(input, "back") {
		if submenu.mode == 0 {
			player.setMenuLogin()
			player.submenu = nil
			return true
		}

		submenu.setMode(gamestate, player, submenu.mode - 1)
	}

	switch submenu.mode {
		case SUBMENU_CREATE_MODE_NAME: {

		}
	}

	return true
}

func (submenu *SubmenuCreate) setMode(gamestate *GameState, player *Player, mode SubmenuCreateMode) {
	submenu.mode = mode
	switch submenu.mode {
		case SUBMENU_CREATE_MODE_NAME: {
			player.send("What is your character's name?")
		}

		case SUBMENU_CREATE_MODE_RACE: {
			player.send("What is your character's race?")
			player.send("The races are:")
			for _, raceData := range gamestate.world.RaceData {
				player.send(fmt.Sprintf("<y>%s</y> - %s", raceData.Name, raceData.Description))
			}
		}

		case SUBMENU_CREATE_MODE_CLASS: {
			player.send("What is your character's class?")
			player.send("The classes are:")
			for _, classData := range gamestate.world.ClassData {
				player.send(fmt.Sprintf("<y>%s</y> - %s", classData.Name, classData.Description))
			}
		}

		case SUBMENU_CREATE_MODE_SUMMARY: {
			// TODO: replace with pretty tables?
			player.send("Here is your character:")
			player.send(fmt.Sprintf("Name: <y>%s</y>", submenu.characterName))
			player.send(fmt.Sprintf("Race: <y>%s</y>", submenu.characterRace))
			player.send(fmt.Sprintf("Class: <y>%s</y>", submenu.characterClass))
			player.send("")
			player.send("Type 'finish' to create your character or 'back' to change something.")
		}
	}
}
