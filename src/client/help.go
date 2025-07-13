package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *uiModel) showHelp() tea.Cmd {
	return func() tea.Msg {
		helpText := "\nAvailable Commands:\n" +
			"  sel <character>     - Select a character\n" +
			"  act <target>        - Set character action on target\n" +
			"  idle                - Set character to idle\n" +
			"  sense               - Sense current area\n" +
			"  inv                 - View character inventory\n" +
			"  drop <item> <qty>   - Drop items from inventory\n" +
			"  say <message>       - Send chat message\n" +
			"  newchar <name>      - Create new character\n" +
			"  echo <text>         - Echo text\n" +
			"  ? <command>         - Get help for specific command\n" +
			"  ?                   - Show this help menu"
		
		return apiResMsg{Green, helpText}
	}
}

func (m *uiModel) showCommandHelp(command string) tea.Cmd {
	return func() tea.Msg {
		var helpText string
		
		switch command {
		case "sel":
			helpText = "\nSelect Character:\n" +
				"Usage: sel <character_name>\n" +
				"Selects a character for other commands to operate on.\n" +
				"You must select a character before using act, sense, inv, drop, say, or idle."
		case "act":
			helpText = "\nSet Action:\n" +
				"Usage: act <target> [amount]\n" +
				"Sets your selected character to perform an action on the specified target.\n" +
				"Targets are resource nodes at your current location (e.g., 'tree', 'rock').\n" +
				"Optional amount parameter limits how many resources to gather before going idle.\n" +
				"Use 'sense' to see available targets."
		case "idle":
			helpText = "\nSet Idle:\n" +
				"Usage: idle\n" +
				"Sets your selected character to idle state, stopping any current action.\n" +
				"Useful for manually stopping resource gathering or other activities."
		case "sense":
			helpText = "\nSense Area:\n" +
				"Usage: sense\n" +
				"Shows characters and resource nodes at your selected character's location.\n" +
				"Use this to find available targets for the 'act' command."
		case "inv":
			helpText = "\nView Inventory:\n" +
				"Usage: inv\n" +
				"Displays your selected character's inventory with items, quantities,\n" +
				"current weight, and total capacity."
		case "drop":
			helpText = "\nDrop Items:\n" +
				"Usage: drop <item_name> [quantity]\n" +
				"Drops items from your inventory. If quantity is omitted, drops all items of that type.\n" +
				"Examples:\n" +
				"  drop wood 5    - drops 5 wood\n" +
				"  drop wood      - drops all wood\n" +
				"  drop balsa logs 2 - drops 2 balsa logs\n" +
				"  drop balsa logs   - drops all balsa logs\n" +
				"Case insensitive. Reduces inventory weight."
		case "say":
			helpText = "\nSend Chat Message:\n" +
				"Usage: say <message>\n" +
				"Sends a chat message as your selected character.\n" +
				"Message appears as '[CharacterName Surname]: message' to all users."
		case "newchar":
			helpText = "\nCreate Character:\n" +
				"Usage: newchar <name>\n" +
				"Creates a new character with the specified name.\n" +
				"Character starts at the default location with empty inventory."
		case "echo":
			helpText = "\nEcho Text:\n" +
				"Usage: echo <text>\n" +
				"Simply echoes back the text you provide. Useful for testing."
		default:
			helpText = fmt.Sprintf("\nNo help available for command '%s'.\nUse '?' to see all available commands.", command)
		}
		
		return apiResMsg{Blue, helpText}
	}
}