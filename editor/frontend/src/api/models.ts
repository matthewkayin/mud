export namespace main {
	
	export class EditorConstants {
	    RoomNone: number;
	    WorldSecondsPerUpdate: number;
	
	    static createFrom(source: any = {}) {
	        return new EditorConstants(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.RoomNone = source["RoomNone"];
	        this.WorldSecondsPerUpdate = source["WorldSecondsPerUpdate"];
	    }
	}
	export class EditorWorldRoom {
	    Position: world.RoomEditorPosition;
	    Room: world.Room;
	    Npcs: world.Npc[];
	    Connections: world.RoomEditorPosition[];
	
	    static createFrom(source: any = {}) {
	        return new EditorWorldRoom(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Position = this.convertValues(source["Position"], world.RoomEditorPosition);
	        this.Room = this.convertValues(source["Room"], world.Room);
	        this.Npcs = this.convertValues(source["Npcs"], world.Npc);
	        this.Connections = this.convertValues(source["Connections"], world.RoomEditorPosition);
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
	export class EditorWorld {
	    Rooms: EditorWorldRoom[];
	    StartRoom?: world.RoomEditorPosition;
	
	    static createFrom(source: any = {}) {
	        return new EditorWorld(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Rooms = this.convertValues(source["Rooms"], EditorWorldRoom);
	        this.StartRoom = this.convertValues(source["StartRoom"], world.RoomEditorPosition);
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

export namespace world {
	
	export enum ChestType {
	    CHEST = 0,
	    CORPSE = 1,
	}
	export enum Direction {
	    COUNT = 4,
	    EAST = 1,
	    NORTH = 0,
	    SOUTH = 2,
	    WEST = 3,
	}
	export enum NpcDisposition {
	    HOSTILE = 1,
	    NEUTRAL = 0,
	}
	export enum NpcMovementType {
	    SENTINEL = 0,
	    WANDER = 1,
	}
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
	export class Item {
	    Id: string;
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
	    ItemId: string;
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
	    Type: ChestType;
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
	
	
	
	
	
	export class ItemData {
	    Name: string;
	    Description: string;
	    Kind: number;
	    Size: number;
	    Data: any;
	
	    static createFrom(source: any = {}) {
	        return new ItemData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.Kind = source["Kind"];
	        this.Size = source["Size"];
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
	    StartingDisposition: NpcDisposition;
	    MovementType: NpcMovementType;
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
	export class RoomEditorPosition {
	    X: number;
	    Y: number;
	
	    static createFrom(source: any = {}) {
	        return new RoomEditorPosition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.X = source["X"];
	        this.Y = source["Y"];
	    }
	}
	export class Room {
	    Name: string;
	    Description: string;
	    IsSafeZone: boolean;
	    EditorPosition: RoomEditorPosition;
	    Exits: number[];
	    ExitIsLockedOnReset: boolean[];
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
	        this.IsSafeZone = source["IsSafeZone"];
	        this.EditorPosition = this.convertValues(source["EditorPosition"], RoomEditorPosition);
	        this.Exits = source["Exits"];
	        this.ExitIsLockedOnReset = source["ExitIsLockedOnReset"];
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

}

