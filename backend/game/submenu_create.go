package game

import (
	"fmt"
	"mud/util"
	"mud/world"
	"strings"
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

	characterSheet world.CharacterSheet
}

func NewSubmenuCreate(gamestate *GameState, player *Player) *SubmenuCreate {
	submenu := &SubmenuCreate {
		mode: SUBMENU_CREATE_MODE_NAME,
		characterSheet: world.CharacterSheet{},
	}

	submenu.setMode(gamestate, player, submenu.mode)
	return submenu
}

func (submenu *SubmenuCreate) handleCommand(gamestate *GameState, player *Player, input string) bool {
	if strings.EqualFold(input, "back") {
		if submenu.mode == 0 {
			player.setMenuLogin(gamestate)
			player.submenu = nil
			return true
		}

		submenu.setMode(gamestate, player, submenu.mode - 1)
		return true
	}

	switch submenu.mode {
		case SUBMENU_CREATE_MODE_NAME: {
			// Validate the player's name
			name, err := world.CharacterNameValidate(gamestate.world, input)
			if err != nil {
				player.send(err.Error())
				return true
			}

			submenu.characterSheet.Name = name
			submenu.setMode(gamestate, player, SUBMENU_CREATE_MODE_RACE)
			return true
		}

		case SUBMENU_CREATE_MODE_RACE: {
			// Search for a race matching the input
			race, ok := gamestate.world.RaceData[strings.ToLower(input)]
			if !ok {
				player.send(fmt.Sprintf("'%s' is not a valid race.", input))
				return true
			}

			submenu.characterSheet.Race = race
			submenu.setMode(gamestate, player, SUBMENU_CREATE_MODE_CLASS)
			return true
		}

		case SUBMENU_CREATE_MODE_CLASS: {
			// Search for a class matching the input
			class, ok := gamestate.world.ClassData[strings.ToLower(input)]
			if !ok {
				player.send(fmt.Sprintf("'%s' is not a valid class.", input))
				return true
			}

			submenu.characterSheet.Class = class
			submenu.setMode(gamestate, player, SUBMENU_CREATE_MODE_SUMMARY)
			return true
		}

		case SUBMENU_CREATE_MODE_SUMMARY: {
			if strings.EqualFold(input, "finish") {
				// Check once more that the character name is available
				_, nameIsTaken := gamestate.world.GetCharacterIfExists(submenu.characterSheet.Name)
				if nameIsTaken {
					player.send(fmt.Sprintf("The character name '%s' has been taken!", submenu.characterSheet.Name))
					submenu.setMode(gamestate, player, SUBMENU_CREATE_MODE_NAME)
					return true
				}

				// Put the character into the characters list
				character := gamestate.world.NewCharacter(player.key, &submenu.characterSheet)
				gamestate.world.AddCharacter(player.key, character)
				gamestate.world.SaveCharacter(character)

				player.send(fmt.Sprintf("Your character has been created. Type 'login %s' to login to them.", character.Name))

				player.submenu = nil
				player.setMenuLogin(gamestate)

				return true
			}

			player.send("'%s' is not a valid command. Type 'finish' or 'back'.")
			return true
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
			player.send("\n")
			player.send("=== RACES ===")
			player.send("\n")

			for _, raceData := range gamestate.world.RaceData {
				player.send(fmt.Sprintf("[%s]", strings.ToUpper(raceData.Name)))
				player.send(fmt.Sprintf("Description : %s", raceData.Description))
				player.send(fmt.Sprintf("Attributes  : %s", raceData.Stats.ToString(true)))
				player.send(fmt.Sprintf("Skills      : %s", util.CombineNames(raceData.Skills.ToStringArray())))
				player.send("\n")
			}

			player.send("What is your character's race?")
		}

		case SUBMENU_CREATE_MODE_CLASS: {
			player.send("\n")
			player.send("=== CLASSES ===")
			player.send("\n")

			for _, classData := range gamestate.world.ClassData {
				player.send(fmt.Sprintf("[%s]", strings.ToUpper(classData.Name)))
				player.send(fmt.Sprintf("Description : %s", classData.Description))
				player.send(fmt.Sprintf("Attributes  : %s", classData.Stats.ToString(true)))
				player.send(fmt.Sprintf("Skills      : %s", util.CombineNames(classData.Skills.ToStringArray())))
				player.send("\n")
			}

			player.send("What is your character's class?")
		}

		case SUBMENU_CREATE_MODE_SUMMARY: {
			stats := submenu.characterSheet.Race.Stats.Add(&submenu.characterSheet.Class.Stats)
			skills := submenu.characterSheet.Race.Skills | submenu.characterSheet.Class.Skills

			player.send("=== CHARACTER SHEET ===")
			player.send("\n")
			player.send(fmt.Sprintf("Name       : <y>%s</y>", submenu.characterSheet.Name))
			player.send(fmt.Sprintf("Race       : %s", submenu.characterSheet.Race.Name))
			player.send(fmt.Sprintf("Class      : %s", submenu.characterSheet.Class.Name))
			player.send(fmt.Sprintf("Attributes : %s", stats.ToString(false)))
			player.send(fmt.Sprintf("Skills     : %s", util.CombineNames(skills.ToStringArray())))
			player.send("\n")
			player.send("Type 'finish' to create your character or 'back' to change something.")
		}
	}
}
