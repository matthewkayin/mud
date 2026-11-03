import { world } from '../../api/models';
import { DIRECTION_NAMES } from '../../store';

// Maps each behavior param name that an NPC or fixture script declares to its type
export type BehaviorParamTypes = Record<string, world.BehaviorParamType>;

// A direction param holds the direction's name
export type BehaviorParamValue = string | number | boolean | world.Item;
export type BehaviorParams = Record<string, BehaviorParamValue>;

// Returns a value for each declared behavior param
export function defaultBehaviorParams(paramTypes: BehaviorParamTypes, itemData: world.ItemData[]): BehaviorParams {
  const params: BehaviorParams = {};
  for (const [name, paramType] of Object.entries(paramTypes)) {
    switch (paramType) {
      case world.BehaviorParamType.STRING:
        params[name] = '';
        break;
      case world.BehaviorParamType.NUMBER:
        params[name] = 0;
        break;
      case world.BehaviorParamType.BOOLEAN:
        params[name] = false;
        break;
      case world.BehaviorParamType.ITEM:
        params[name] = world.Item.createFrom({
          Id: itemData.length > 0 ? itemData[0].Name : '',
          Amount: 1,
          Durability: 0,
        });
        break;
      case world.BehaviorParamType.DIRECTION:
        params[name] = DIRECTION_NAMES[world.Direction.NORTH];
        break;
    }
  }

  return params;
}
