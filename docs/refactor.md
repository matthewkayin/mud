# Refactor Survival Guide

This is a guide explaining what has changed in the refactor, or at least some of the big ideas.

- Usernames and passwords are now used to login and playerId has been replaced with playerKey
- User accounts are stored as flat file JSON in saves/accounts. Each account JSON contains only username and a hash of their password
- There is a "database" folder for saving JSON files. This folder is not actually the database, but it is our persistance layer. It provides a helper that lets us save a JSON atomically.
  - This is necessary to prevent partial-saves from corrupting a file. If a JSON save is interrupted, the old file is overwritten only halfway. The database package fixes this by first saving a ".json.tmp" version of the file, and then copying the file over once the save is successful (moving the file is an atomic operation on posix filesystems)
- Menus in gamestate no longer have an onEnter and onExit function, instead menu transitions are handled by player.setMenuLogin() and player.setMenuWorld()
- Menus now have support for help "topics" which can be accessed by the help command. 
  - Also, every menu command automatically inserts its own "topic" during menu initialization, which means that anything we query via the "help" command is just a topic. 
  - A topic is basically just a list of strings to print to the user when they type "help <topic>"
  - All topics (and commands) can have one or more categories. If you type "help <category>" it shows you all the topics in the category
  - Since "help" matches on topics, categories, and commands, this means that all three of these things must have distinct names, and the server will yell at you on startup if you duplicate names
  - Matching on topics is done with lowercase of the user input, so you should not make capitalized topics, although I suppose we could fix this by doing strings.EqualFold() if we wanted capital letters for each Topic / Category
- There is now a concept called Submenus. 
  - A submenu is a menu that players can go into. Players may only have one submenu at a time.
  - If the player submenu is nil, then the submenu obviously does nothing
  - If the player submenu is not-nil, then before we run a command through the menu entries, we will first run it through the submenu
  - The submenu does not have "entries" in the way that a menu does. Instead, it just has a general handleCommand() function. This allows it handle arbitrary input like "Yes" and "No", in addition to commands
  - Each submenu has its own arbitrary data, so for example, the character create submenu has a character sheet as well as a number which tells us where we are in the creation flow
  - The submenu's handleCommand() returns a bool indicating whether it handled the command. 
    - If the submenu returns true, then the buck stops there, and no further command handling is done in the menu
    - If the submenu returns false, then the input flows down into the menu and we try to call on one of the commands in the menu's entries as normal
  - This means is that submenus can serve two purposes. 
    - You can have a submenu that completely takes over the parent menu (such as character create). 
    - Or you can have a submenu that tries to intercept commands but still allows normal command usage (such as NPC dialog, sign boards, etc.)
- I'm doing weird things with MobData (specifically, struct embedding, which is a Go feature, read more here: https://gobyexample.com/struct-embedding)
 - Struct embedding means that all the fields and member functions on MobData are now available natively on the Mob and Character, so we can just say Mob.Name and Character.Name instead of Mob.Data.Name and Character.Data.Name
 - Additionally, the Mob's MobData is now a pointer to a MobData, and on a player-controlled mob, it points back to the Character's MobData, so any changes made on the Mob are automatically saved and written back to the source Character
- Rooms are now Lua scripts
- All the script data that was previously in global arrays has been moved into the World struct. For example CLASS_DATA and RACE_DATA are now world.ClassData and world.RaceData
  - This may prove inconvenient later when we want to have items reference something in ITEM_DATA, we'll see how it shakes out as I add more features
  - One of the reasons this is done is to make it easier to support hot-reloading of Lua scripts when using the editor
- Mob movement is now on the mob itself instead of the room
