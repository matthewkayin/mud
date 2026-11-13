local npc = {}

npc.id = "happy_cow"
npc.name = "Happy Cow"
npc.description = "Lives a life of total bovine simplicity."

npc.experience_worth = 10

npc.max_health = 10
npc.damage = 2
npc.starting_disposition = world.NpcDisposition.NEUTRAL

npc.drop_table = {
    {
        item = "Leather",
        amount = 1,
        drop_chance_percent = 100,
    }
}

function npc.init(self)
    return {
        moo_timer = world.time_seconds(15)
    }
end

function npc.update(self, instance)
    instance.moo_timer = instance.moo_timer - 1
    if instance.moo_timer < 0 then
        local room = world.mob_get_room(self)
        local message = math.random(2)
        if message == 1 then
            world.message_room(room, "The happy cow chews some grass.")
        else
            world.message_room(room, "The cow makes a mooing sound.")
        end

        instance.moo_timer = world.time_seconds(15)
    end
end

return npc
