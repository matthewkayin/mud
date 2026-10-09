local npc = {}

npc.id = "troll"
npc.name = "Troll"
npc.description = "A hairy beast with a large nose and a pallid complexion towers over you."

npc.experience_worth = 100
npc.experience_worth_scaling = 25

npc.stats = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.scaling = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.NEUTRAL
npc.movement_type = world.NpcMovementType.SENTINEL

-- The troll blocks an exit until a player pays it the toll
npc.behavior_params = {
	{ name = "Toll", type = world.NpcBehaviorParamType.ITEM },
	{ name = "ExitToBlock", type = world.NpcBehaviorParamType.DIRECTION },
}

npc.init = function(params)
	-- TODO: block the exit, demand the toll from entering players, and step aside once
	-- it is paid. This needs script API functions for exits, messages and inventories.
	return {
		toll = params.Toll,
		exit_to_block = params.ExitToBlock,
	}
end

return npc
