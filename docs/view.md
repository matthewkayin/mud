# View Panel

The bottom-right panel is currently the inventory panel. This panel should be changed to being a multi-purpose "view" panel that can be configured according to the users liking. The panel just displays text and has no buttons or interactive components.

The panel has the following modes: 
- Help - An introductory mode that explains what the view panel is and how to use it
- Chat - A mode that shows a chatlog
- Inventory - A mode that shows the users inventory
- Equipment - A mode that shows the users current equipment

Users interact with this panel using a new command called "view <mode>". This allows them to change how the view panel is configured from the frontend. Typing "view inventory" will therefore show the inventory in the view panel. Typing "view chat" will view the chat.

## Technical Considerations

Generally the contents of the view panel are small enough that they can be sent in a heartbeat as requested by the user. If the user is viewing the inventory, for example, and then they type "view equipment", it is trivial for the backend to send a new set of lines to the user to be displayed on the equipment panel. However, it may not be trivial to send the entire chat history to the user. The chat displayed in the view panel should not just be a few lines, it should be a scrollable history of all chat messages. The frontend should consider any socket message that is prefixed with square bracket tags (i.e. "[CHAT] This is a chat message" or "[WHISPER] This is also a chat message" or "[ANNOUNCEMENT] This is a server announcement") to be a chat message. These chat messages should be kept in state by the frontend, and when users choose to "view chat" the frontend should simply switch to displaying the already-stored chat lines in the view panel. It is up to the implementation whether "view chat" skips the server round trip. "view chat" could be "intercepted" by the frontend and handled on the frontend only. Or "view chat" could dispatch to the server and the server could send back a response which indicates to the frontend to change its view mode to chat. If the frontend "interception" is valuable, then it may be the case that the frontend should intercept "view" commands for all modes, and that the frontend should store the lines for all modes in state, and that the backend should always keep the frontend informed through updates, in other words, the backend will always send up-to-date inventory and equipment whenever those things change.

## Inventory Formatting

The "view inventory" panel shall have the following formatting. Note that the "inventory" header below might not be perfectly center aligned, but the one in the code should be. The "Size" column should always be wide enough to support the string "Size (100 / 100)" and the "Amount" column should always be wide enough to support the string "Amount". The "Item" column should take up the rest of the remaining space, and it should never shrink even if the items in the inventory have short names. The item column should have some maximum acceptable item name length. Here I have chosen "Damaged Sword of the Forgotten Kingdom" as an extreme example. If an item name is longer than the max name, then the view panel should truncate the name with a trailing "...". Alternatively, it may be acceptable to just have all the columns grow as-needed and to accept horizontal scrollbars in the case of long names, since I have observed that players who are hard-of-sight tend to zoom in the page on the browser in order to increase the font size.

The formatted table described here could be built by the backend, or there is a case to be made that the backend should send an array of item JSONs containing name, amount, and size, and then the frontend should build the formatted string.

+--------------------------- Inventory ------------------------------+
| Item                                   | Amount | Size (100 / 100) |
|----------------------------------------|--------|------------------|
| Damaged Sword of the Forgotten Kingdom | x1     | 10               |
+--------------------------------------------------------------------+
