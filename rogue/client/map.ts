import {Pos,Rect,Log} from "./util";
import {TileSheet,TileIndex} from "./tile";
import {Grid} from "./grid";

// Map defines the layout of a dungeon.
//   w = number of cells horizontally
//   h = number of cells vertically
//
//   rows = string[][] terrain definitions per cell.
//   Each cell is composed of a terrain type and a tile index.
//   Terrain type options are:
//      "."  floor
//      "#"  wall
//      "+"  closed door
//      "/"  open door
//   Tile index points to the tilesheet tile graphic to render in the cell
//   which can be a single tile index number, tile [row,col] coordinates,
//   or the tile alias string. Examples below:
//
//      "92"   tile index number
//      "5,12" tile row,col
//      "W2"   tile alias
//
//   The terrain type and tile index are concatenated together.
//
//   Example map rows:
//   [".1", ".1", ".2", "#3", "#4", "+5"]              (tile indexes)
//   [".0,1", ".0,1", ".0,2", "#0,3", "#0,4", "+0,5"]  (tile row,col)
//   [".F1", ".F1", ".F2", "#W1", "#W2", "+D1"]        (tile aliases)
//   
//

interface ItemLoc {
  ts: TileSheet,
  ti: TileIndex,
  pos: Pos,
}

interface Map {
  grid: Grid,
  ts: TileSheet,
  w: number,
  h: number,
  terrain: string,
  items: ItemLoc[],
}

function NewMap(grid:Grid, ts:TileSheet, terrain:string, items:ItemLoc[]):Map {
  const map = <Map>{};
  map.grid = grid;
  map.ts = ts;
  map.terrain = terrain;
  map.items = items;

  // Discard blank rows and extra whitespace in cols when counting.

  const rows = terrain.split("\n").filter(row => row.trim().length > 0);
  map.h = rows.length;

  if (rows.length > 0) {
    const cols = rows[0].split(" ").filter(col => col.trim().length > 0);
    map.w = cols.length;
  } else {
    map.w = 0;
  }

  return map;
}


// Draw map to canvas ctx starting from top left offset position in map.
function MapDraw(map:Map) {
  // Draw terrain
  //
  // Discard empy lines and extra whitespace in rows and cols.
  const rows = map.terrain.split("\n").filter(row => row.trim().length > 0);
  for (let yrow=0; yrow < rows.length; yrow++) {
    const row = rows[yrow];
    const cols = row.split(" ").filter(col => col.trim().length > 0);
    for (let xcol=0; xcol < cols.length; xcol++) {
      const scell = cols[xcol];

      // Not long enough to contain both terrain type and tile index.
      if (scell.length <= 1) {
        break;
      }

      const sTerrain = scell[0];
      const sTileIndex = scell.substring(1);

      let ti:TileIndex;

      // tile index number?
      if (!isNaN(sTileIndex as any)) {
        ti = parseInt(sTileIndex);
        if (ti === NaN) {
          Log(`Illegal tile index: '${sTileIndex}'`);
          break;
        }
      // tile row,col?
      } else if (sTileIndex.indexOf(",") != -1) {
        const [srow, scol] = sTileIndex.split(",");
        const row = parseInt(srow);
        const col = parseInt(scol);
        if (row === NaN || col === NaN) {
          Log(`Illegal tile index: '${sTileIndex}'`);
          break;
        }
        ti = [row,col];
      // tile alias?
      } else {
        ti = sTileIndex;
      }

      map.grid.DrawTile(map.ts, ti, xcol, yrow);
    }
  }

  // Draw items
  for (const itemLoc of map.items) {
    map.grid.DrawTile(map.ts, itemLoc.ti, itemLoc.pos.x, itemLoc.pos.y);
  }
}

export {
  ItemLoc,
  Map,
  NewMap,
  MapDraw,
};
