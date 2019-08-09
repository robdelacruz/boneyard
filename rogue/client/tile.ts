import {Rect,PosT} from "./util";

// Allow mnemonic aliases to be mapped to a tile index.
// Ex. {"F1": 3, "F2": 4, "W1": 5, "W2": 6} for floor tiles and wall tiles.
type TileAliases = {[alias:string]:number|PosT};

// Union type specifying either index number, row/col, or alias for
// specifying tile indexes.
type TileIndex = number | PosT | string;

class TileSheet {
  img: HTMLImageElement;
  wcell: number;
  hcell: number;
  margin: number;

  nRows: number;
  nCols: number;
  nTiles: number;
  tileAliases: TileAliases;

  constructor(
    img:HTMLImageElement,
    wcell:number,
    hcell:number,
    margin:number=0,
    tileAliases={}
  ) {
    this.img = img;
    this.wcell = wcell;
    this.hcell = hcell;
    this.margin = margin;
    this.tileAliases = tileAliases;

    this.nCols = Math.floor((img.width + margin*2) / (wcell + margin));
    this.nRows = Math.floor((img.height + margin*2) / (hcell + margin));
    this.nTiles = this.nCols * this.nRows;
  }

  // Return rect bounds of tile referenced by tile index.
  TileRect(ti:TileIndex) {
    let iTile = 0;
    if (typeof ti === "number") {
      iTile = ti;
    } else if (typeof ti === "object") {
      iTile = this.TileNFromPos(ti);
    } else if (typeof ti === "string") {
      iTile = this.TileNFromAlias(ti);
    }

    iTile = iTile % this.nTiles;
    const row = Math.floor(iTile / this.nCols);
    const col = iTile % this.nCols;

    const rect = <Rect>{
      x: col * (this.wcell + this.margin),
      y: row * (this.hcell + this.margin),
      w: this.wcell,
      h: this.hcell,
    };
    return rect;
  }

  // Return tile index given row and column in tilesheet.
  TileNFromPos(posTile:PosT):number {
    let [row, col] = posTile;
    row = row % this.nRows;
    col = col % this.nCols;
    return row * this.nCols + col;
  }

  TileNFromAlias(alias:string):number {
    // Alias not defined, return first tile.
    if (this.tileAliases[alias] == null) {
      return 0;
    }
    const ti:TileIndex = this.tileAliases[alias];
    if (typeof ti === "object") {
      return this.TileNFromPos(ti);
    }
    return ti;
  }
}

export {
  TileSheet,
  TileIndex,
};


