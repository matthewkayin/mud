## Example Character Sheet

+-----------------------------------------+
| Character Name                          |
|-----------------------------------------|
| Race  : Gremlin                         |
| Class : Priest                          |
| Job   : Alchemist                       |
|-----------------------------------------|
| Level      : 1                          |
| Experience : 0 / 300                    |
|-----------------------------------------|
| HP : 30 / 30                            | 
| MP : 30 / 30                            |
|-----------------------------------------|
| Stat         | Base | Equipment | Total |
|-----------------------------------------|
| Vitality     | 10   |    ---    | 10    |
| Strength     | 12   |    ---    | 12    |
| Agility      | 6    |    ---    | 6     |
| Intelligence | 5    |    ---    | 5     |
| Faith        | 8    |    ---    | 8     |
+-----------------------------------------+

This layout is drawn by `printSheetTable` in `backend/game/menu_utils.go`, which the `stats` command and the character creator's sheet both use. It takes a title, a list of sections and a table. Each section is a flat `key, value, key, value...` slice, and the `:` is aligned within each section separately. The table is as wide as its widest part, and if a title or section line is wider than the stat table, the table's last column is widened to match. Plain tables with no key/value sections, such as inventory and spells, use `printTable`.
