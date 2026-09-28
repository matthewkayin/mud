export type Vec2 = {
  x: number;
  y: number;
}

export type Rect = {
  x: number;
  y: number;
  width: number;
  height: number;
}

export function rectHasPoint(rect: Rect, point: Vec2): boolean {
  return !(
    point.x < rect.x ||
    point.y < rect.y ||
    point.x >= rect.x + rect.width ||
    point.y >= rect.y + rect.height
  );
}

export function vec2FromDirection(direction: number): Vec2 {
  switch (direction) {
    case 0:
      return { x: 0, y: -1 };
    case 1:
      return { x: 1, y: 0 };
    case 2:
      return { x: 0, y: 1 };
    case 3:
      return { x: -1, y: 0 };
    default:
      return { x: 0, y: 0 };
  }
}
