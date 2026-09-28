export type Vec2 = {
  x: number;
  y: number;
}

export function vec2Add(a: Vec2, b: Vec2): Vec2 {
  return {
    x: a.x + b.x,
    y: a.y + b.y
  };
}

export function vec2Subtract(a: Vec2, b: Vec2): Vec2 {
  return {
    x: a.x - b.x,
    y: a.y - b.y
  };
}
