export {
  Pos,
  Rect,
  PosT,
  RectT,
  Log,
};

interface Pos {
  x: number;
  y: number;
}

interface Rect {
  x: number;
  y: number;
  w: number;
  h: number;
}

type PosT=[number, number];
type RectT=[number, number, number, number];

function Log(s: string) {
  console.log(s);
}

