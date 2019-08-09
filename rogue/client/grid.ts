// Grid

import {Pos,Rect} from "./util";
import {Sprite} from "./sprite";
import {TileSheet,TileIndex} from "./tile";

class Grid {
  cv: HTMLCanvasElement;
  ctx: CanvasRenderingContext2D;
  wcell: number;

  constructor(cv:HTMLCanvasElement, wcell:number) {
    this.wcell = wcell;
    this.cv = cv;
    this.ctx = cv.getContext("2d")!;
  }

  CellPos(xcell:number, ycell:number):Pos {
    return <Pos>{
      x: xcell * this.wcell,
      y: ycell * this.wcell,
    };
  }

  CellRect(xcell:number, ycell:number):Rect {
    return <Rect>{
      x: xcell * this.wcell,
      y: ycell * this.wcell,
      w: this.wcell,
      h: this.wcell,
    };
  }

  DrawSprite(sp:Sprite, xcell:number, ycell:number) {
    this.drawObj(sp.img, sp.FrameRect(), xcell, ycell);
  }

  DrawTile(ts:TileSheet, ti:TileIndex, xcell:number, ycell:number) {
    this.drawObj(ts.img, ts.TileRect(ti), xcell, ycell);
  }

  drawObj(img:HTMLImageElement, srcRect:Rect, xcell:number, ycell:number) {
    const {x,y} = this.CellPos(xcell, ycell);
    this.ctx.drawImage(
      img,
      srcRect.x, srcRect.y,
      srcRect.w, srcRect.h,
      x,y,
      this.wcell, this.wcell,
    );
  }
}

export {
  Grid,
};
