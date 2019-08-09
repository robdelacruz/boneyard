import {Rect} from "./util";
import {TileSheet,TileIndex} from "./tile";

class Sprite {
  ts: TileSheet;
  ti: TileIndex;
  img: HTMLImageElement;
  x: number;
  y: number;
  props: {[id:string]:string};

  constructor(ts:TileSheet, ti:TileIndex) {
    this.ts = ts;
    this.ti = ti;
    this.img = ts.img;
    this.x = 0;
    this.y = 0;
    this.props = {};
  }

  FrameRect():Rect {
    return this.ts.TileRect(this.ti);
  }
}

export {
  Sprite,
};

