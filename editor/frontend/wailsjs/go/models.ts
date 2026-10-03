export namespace world {
	
	export class Behavior {
	    Hooks: any;
	
	    static createFrom(source: any = {}) {
	        return new Behavior(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Hooks = source["Hooks"];
	    }
	}
	export class Equipment {
	    IsSlotInUse: boolean[];
	    SlotItem: Item[];
	    StatBonuses: StatBlock;
	
	    static createFrom(source: any = {}) {
	        return new Equipment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.IsSlotInUse = source["IsSlotInUse"];
	        this.SlotItem = this.convertValues(source["SlotItem"], Item);
	        this.StatBonuses = this.convertValues(source["StatBonuses"], StatBlock);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Item {
	    Id: number;
	    Amount: number;
	    Durability: number;
	
	    static createFrom(source: any = {}) {
	        return new Item(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Amount = source["Amount"];
	        this.Durability = source["Durability"];
	    }
	}
	export class Inventory {
	    Items: Item[];
	
	    static createFrom(source: any = {}) {
	        return new Inventory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Items = this.convertValues(source["Items"], Item);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StatBlock {
	    Vitality: number;
	    Strength: number;
	    Agility: number;
	    Intelligence: number;
	    Faith: number;
	
	    static createFrom(source: any = {}) {
	        return new StatBlock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Vitality = source["Vitality"];
	        this.Strength = source["Strength"];
	        this.Agility = source["Agility"];
	        this.Intelligence = source["Intelligence"];
	        this.Faith = source["Faith"];
	    }
	}
	export class MobData {
	    Name: string;
	    Room: number;
	    Level: number;
	    Experience: number;
	    ExperienceToNextLevel: number;
	    Stats: StatBlock;
	    Abilities: number;
	    Health: number;
	    Mana: number;
	    Spells: number[];
	    Inventory: Inventory;
	    Equipment: Equipment;
	
	    static createFrom(source: any = {}) {
	        return new MobData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Room = source["Room"];
	        this.Level = source["Level"];
	        this.Experience = source["Experience"];
	        this.ExperienceToNextLevel = source["ExperienceToNextLevel"];
	        this.Stats = this.convertValues(source["Stats"], StatBlock);
	        this.Abilities = source["Abilities"];
	        this.Health = source["Health"];
	        this.Mana = source["Mana"];
	        this.Spells = source["Spells"];
	        this.Inventory = this.convertValues(source["Inventory"], Inventory);
	        this.Equipment = this.convertValues(source["Equipment"], Equipment);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CharacterEquippedSpell {
	    EquipCount: number;
	    Casts: number;
	    IsKnown: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CharacterEquippedSpell(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.EquipCount = source["EquipCount"];
	        this.Casts = source["Casts"];
	        this.IsKnown = source["IsKnown"];
	    }
	}
	export class Character {
	    PlayerId: number;
	    Race: number;
	    Class: number;
	    Job: number;
	    SpellsEquipped: Record<number, CharacterEquippedSpell>;
	    SpellsKnown: number[];
	    ClassSpells: number[];
	    RecipesKnown: number[];
	    RoomsDiscovered: number[];
	    Data: MobData;
	
	    static createFrom(source: any = {}) {
	        return new Character(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PlayerId = source["PlayerId"];
	        this.Race = source["Race"];
	        this.Class = source["Class"];
	        this.Job = source["Job"];
	        this.SpellsEquipped = this.convertValues(source["SpellsEquipped"], CharacterEquippedSpell, true);
	        this.SpellsKnown = source["SpellsKnown"];
	        this.ClassSpells = source["ClassSpells"];
	        this.RecipesKnown = source["RecipesKnown"];
	        this.RoomsDiscovered = source["RoomsDiscovered"];
	        this.Data = this.convertValues(source["Data"], MobData);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Int32Range {
	    Min: number;
	    Max: number;
	
	    static createFrom(source: any = {}) {
	        return new Int32Range(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Min = source["Min"];
	        this.Max = source["Max"];
	    }
	}
	export class DropTableEntry {
	    ItemId: number;
	    AmountRange: Int32Range;
	    DurabilityPercentRange: Int32Range;
	    DropChancePercent: number;
	
	    static createFrom(source: any = {}) {
	        return new DropTableEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ItemId = source["ItemId"];
	        this.AmountRange = this.convertValues(source["AmountRange"], Int32Range);
	        this.DurabilityPercentRange = this.convertValues(source["DurabilityPercentRange"], Int32Range);
	        this.DropChancePercent = source["DropChancePercent"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DropTable {
	    Entries: DropTableEntry[];
	
	    static createFrom(source: any = {}) {
	        return new DropTable(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Entries = this.convertValues(source["Entries"], DropTableEntry);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Chest {
	    Name: string;
	    Type: number;
	    Timer: number;
	    DropTable: DropTable;
	    Inventory: Inventory;
	
	    static createFrom(source: any = {}) {
	        return new Chest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Type = source["Type"];
	        this.Timer = source["Timer"];
	        this.DropTable = this.convertValues(source["DropTable"], DropTable);
	        this.Inventory = this.convertValues(source["Inventory"], Inventory);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class Event {
	    EventType: number;
	    Data: any;
	
	    static createFrom(source: any = {}) {
	        return new Event(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.EventType = source["EventType"];
	        this.Data = source["Data"];
	    }
	}
	
	
	
	export class ItemData {
	    Name: string;
	    Description: string;
	    ItemType: number;
	    Data: any;
	
	    static createFrom(source: any = {}) {
	        return new ItemData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.ItemType = source["ItemType"];
	        this.Data = source["Data"];
	    }
	}
	
	export class MobHandle {
	    Id: number;
	    Generation: number;
	
	    static createFrom(source: any = {}) {
	        return new MobHandle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Generation = source["Generation"];
	    }
	}
	export class Npc {
	    Type: number;
	    LevelRange: Int32Range;
	    StartingDisposition: number;
	    MovementType: number;
	    Behavior: Behavior;
	    SpawnRoom: number;
	    RespawnDuration: number;
	    SleepDuration: number;
	    AwakeDuration: number;
	    MovementStepDuration: number;
	    DropTable: DropTable;
	
	    static createFrom(source: any = {}) {
	        return new Npc(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Type = source["Type"];
	        this.LevelRange = this.convertValues(source["LevelRange"], Int32Range);
	        this.StartingDisposition = source["StartingDisposition"];
	        this.MovementType = source["MovementType"];
	        this.Behavior = this.convertValues(source["Behavior"], Behavior);
	        this.SpawnRoom = source["SpawnRoom"];
	        this.RespawnDuration = source["RespawnDuration"];
	        this.SleepDuration = source["SleepDuration"];
	        this.AwakeDuration = source["AwakeDuration"];
	        this.MovementStepDuration = source["MovementStepDuration"];
	        this.DropTable = this.convertValues(source["DropTable"], DropTable);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class NpcData {
	    Name: string;
	
	    static createFrom(source: any = {}) {
	        return new NpcData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	    }
	}
	export class Room {
	    Name: string;
	    Description: string;
	    Exits: number[];
	    ExitIsLocked: boolean[];
	    IsSafeZone: boolean;
	    DropTable: DropTable;
	    Inventory: Inventory;
	    Chests: Chest[];
	
	    static createFrom(source: any = {}) {
	        return new Room(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.Exits = source["Exits"];
	        this.ExitIsLocked = source["ExitIsLocked"];
	        this.IsSafeZone = source["IsSafeZone"];
	        this.DropTable = this.convertValues(source["DropTable"], DropTable);
	        this.Inventory = this.convertValues(source["Inventory"], Inventory);
	        this.Chests = this.convertValues(source["Chests"], Chest);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class World {
	    Rooms: Room[];
	    Npcs: Npc[];
	
	    static createFrom(source: any = {}) {
	        return new World(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Rooms = this.convertValues(source["Rooms"], Room);
	        this.Npcs = this.convertValues(source["Npcs"], Npc);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

