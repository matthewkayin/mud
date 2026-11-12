package game

import (
	"fmt"
	"mud/util"
	"mud/world"
	"strconv"
	"strings"
)

type SubmenuCreateMode int
const (
	SUBMENU_CREATE_MODE_NAME SubmenuCreateMode = iota
	SUBMENU_CREATE_MODE_RACE
	SUBMENU_CREATE_MODE_BACKGROUND
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
			race, err := gamestate.world.RaceIdFromString(input)
			if err != nil {
				player.send(err.Error())
				return true
			}

			submenu.characterSheet.Race = race
			submenu.setMode(gamestate, player, SUBMENU_CREATE_MODE_BACKGROUND)
			return true
		}

		case SUBMENU_CREATE_MODE_BACKGROUND: {
			// Search for a class matching the input
			class, err := gamestate.world.BackgroundIdFromString(input)
			if err != nil {
				player.send(err.Error())
				return true
			}

			submenu.characterSheet.Background = class
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
				player.send(fmt.Sprintf("Description: %s", raceData.Description))
				player.send(fmt.Sprintf("Attributes:  %s", attributeBonusString(&raceData.Attributes)))
				player.send(fmt.Sprintf("Skills:      %s", skillBonusString(&raceData.Skills)))
				player.send("\n")
			}

			player.send("What is your character's race?")
		}

		case SUBMENU_CREATE_MODE_BACKGROUND: {
			player.send("\n")
			player.send("=== BACKGROUNDS ===")
			player.send("\n")

			for _, backgroundData := range gamestate.world.BackgroundData {
				player.send(fmt.Sprintf("[%s]", strings.ToUpper(backgroundData.Name)))
				player.send(fmt.Sprintf("Description: %s", backgroundData.Description))
				player.send(fmt.Sprintf("Attributes:  %s", attributeBonusString(&backgroundData.Attributes)))
				player.send(fmt.Sprintf("Skills:      %s", skillBonusString(&backgroundData.Skills)))
				player.send("\n")
			}

			player.send("What is your character's background?")
		}

		case SUBMENU_CREATE_MODE_SUMMARY: {
			raceData := gamestate.world.RaceData[submenu.characterSheet.Race]
			backgroundData := gamestate.world.BackgroundData[submenu.characterSheet.Background]
			attributes := raceData.Attributes.Add(&backgroundData.Attributes)
			skills := raceData.Skills.Add(&backgroundData.Skills)

			player.send("=== CHARACTER SHEET ===")
			player.send("\n")
			player.send(fmt.Sprintf("Name:       <y>%s</y>", submenu.characterSheet.Name))
			player.send(fmt.Sprintf("Race:       %s", raceData.Name))
			player.send(fmt.Sprintf("Background: %s", backgroundData.Name))
			player.send(fmt.Sprintf("Attributes: %s", attributeBonusString(&attributes)))
			player.send(fmt.Sprintf("Skills:     %s", skillBonusString(&skills)))
			player.send("\n")
			player.send("Type 'finish' to create your character or 'back' to change something.")
		}
	}
}

func attributeBonusString(attributes *world.AttributeBlock) string {
	var builder strings.Builder
	builder.Grow(8 * world.ATTRIBUTE_COUNT)

	for index := range world.ATTRIBUTE_COUNT {
		builder.WriteString(world.ATTRIBUTE_DATA[index].Abbreviation)
		builder.WriteByte(' ')

		if attributes.Values[index] >= 0 {
			builder.WriteByte('+')
		}
		builder.WriteString(strconv.Itoa(int(attributes.Values[index])))

		if index < world.ATTRIBUTE_COUNT - 1 {
			builder.Write([]byte(" | "))
		}
	}

	return builder.String()
}

func skillBonusString(skills *world.SkillBlock) string {
	skillBonuses := make([]string, 0, world.SKILL_COUNT)

	for index := range world.SKILL_COUNT {
		if skills.Values[index] == 0 {
			continue
		}

		skillBonuses = append(skillBonuses, fmt.Sprintf("%s +%d", world.SKILL_DATA[index].Name, skills.Values[index]))
	}

	return util.CombineNames(skillBonuses)
}
