Implement a new test for durability. The test should check for durability on a weapon, a shield, and a piece of armor. It should have a mock item lua script for each of these items in its test data. It should setup the starting room so that there are four chests, each one containing one each of the three items, with each chest containing different durabilities:
  - damaged (one hit from being broken)
  - damaged
  - half-durability (one hit from being damaged)
  - full-durability
  - reinforced (one hit from losing reinforcement)
  - fully-reinforced

There should be six test functions, one for each chest. During each test, the player should
  - create a character
  - 'take all from <chest>'
  - equip the sword, shield, and armor
  - 'look at <item> in <slot>' and assert that the item name has the correct durability prefix
  - attack the dummy
  - assert for equipment break messages as appropriate
  - finish killing the dummy
  - 'look at <item> in <slot>' and assert that the item name has the correct durability prefix
  - try to 'remove from <slot>' for each item and assert the correct message is shown

Expected outcomes for each scenario

| Scenario | First 'look at' | Break messages | Second 'look at' | 'remove from <slot>' |
| -- | -- | -- | -- | --- |
| Damaged, one hit from broken | Items should have 'Damaged' prefix | Item broken messages should appear for each item | Error message should be reported indicating that no item exists in the slot | Error message should be reported indicating that no item exists in the slot |
| Damaged, half-durability | Items should have 'Damaged' prefix | No break message should appear | Items should have 'Damaged' prefix | Items should be removed with damaged prefix in their name |
| Half-durability, one hit from damaged | Items should have no prefix | Item damaged message should appear for each item | Items should have 'Damaged' prefix | Items should be removed with damaged prefix in their name |
| Full-durability | Items should have no prefix | No break message should appear | Items should have no prefix | Items should be removed with no prefix in their name |
| Reinforced, one hit from losing reinforcement | Items should have reinforced prefix | Lost reinforcement message should appear | Items should have no prefix | Items should be removed with no prefix in their name |
| Fully-reinforced | Items should have reinforced prefix | No break message should appear | Items should have reinforced prefix | Items should be removed with reinforced prefix in their name |
